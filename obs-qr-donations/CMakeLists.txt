cmake_minimum_required(VERSION 3.16)
project(obs-qr-donations VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTORCC ON)
set(CMAKE_AUTOUIC ON)

# Find required packages
find_package(Qt6 COMPONENTS Widgets Svg Network REQUIRED)
find_package(LibOBS REQUIRED)
find_package(obs-frontend-api REQUIRED)

# Breez SDK Spark setup
# Allow forcing the stub or overriding detection for easier local testing/CI
option(BREEZ_USE_STUB "Force using the Breez stub implementation even if the SDK is present" OFF)
option(BREEZ_STUB_SIMULATE "When using the Breez stub, simulate invoice creation and sends for UI testing" ON)
set(BREEZ_SDK_PATH "${CMAKE_CURRENT_SOURCE_DIR}/third_party/breez_sdk" CACHE PATH "Path to Breez SDK Spark")

if(BREEZ_USE_STUB)
    message(WARNING "BREEZ_USE_STUB is ON — forcing use of the Breez stub implementation")
    set(BREEZ_SDK_FOUND FALSE)
else()
    if(EXISTS "${BREEZ_SDK_PATH}/CMakeLists.txt")
        add_subdirectory(${BREEZ_SDK_PATH} breez_sdk)
        set(BREEZ_SDK_FOUND TRUE)
        message(STATUS "Breez SDK Spark found at ${BREEZ_SDK_PATH}")
        add_compile_definitions(HAVE_BREEZ_SDK)
        add_compile_definitions(BREEZ_SDK_SPARK)
    else()
        message(WARNING "========================================")
        message(WARNING "Breez SDK Spark not found at ${BREEZ_SDK_PATH}")
        message(WARNING "Lightning/Spark wallet features will be disabled.")
        message(WARNING "")
        message(WARNING "To enable Spark SDK:")
        message(WARNING "1. Clone the SDK: git clone https://github.com/breez/spark-sdk.git third_party/breez_sdk")
        message(WARNING "2. Run setup script: scripts/setup_spark_sdk.bat (Windows) or scripts/setup_spark_sdk.sh (Linux/Mac)")
        message(WARNING "3. Reconfigure CMake")
        message(WARNING "========================================")
        set(BREEZ_SDK_FOUND FALSE)
    endif()
endif()

if(BREEZ_USE_STUB AND BREEZ_STUB_SIMULATE)
    add_compile_definitions(BREEZ_STUB_SIMULATE)
    message(STATUS "Breez stub simulation enabled (BREEZ_STUB_SIMULATE)")
endif()

# Spark wallet configuration
option(ENABLE_SPARK_WALLET "Enable Spark wallet support" ON)
if(ENABLE_SPARK_WALLET)
    add_definitions(-DENABLE_SPARK_WALLET)
    message(STATUS "Spark wallet support enabled")
endif()

# Include directories
include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}/src
    ${LIBOBS_INCLUDE_DIR}
    ${OBS_FRONTEND_API_INCLUDE_DIR}
)

# Plugin version information
set(PLUGIN_NAME "obs-qr-donations")
set(PLUGIN_VERSION "1.0.0")
set(PLUGIN_DESCRIPTION "OBS Plugin for displaying QR codes to receive cryptocurrency donations")
set(PLUGIN_AUTHOR "Your Name")

# Source files
set(SOURCES
    src/plugin-main.cpp
    src/qr-donations.cpp
    src/qr-donations.hpp
    src/qr-widget.cpp
    src/qr-widget.hpp
    src/qr-generator.cpp
    src/qr-generator.hpp
    src/asset-manager.cpp
    src/asset-manager.hpp
    src/donation-effect.cpp
    src/donation-effect.hpp
)

# UI files
qt6_wrap_ui(UI_HEADERS
    ui/qr-donations.ui
)

# Resources
qt6_add_resources(RESOURCES
    resources/resources.qrc
)

# Create the plugin library
add_library(${PLUGIN_NAME} MODULE
    ${SOURCES}
    ${UI_HEADERS}
    ${RESOURCES}
)

# Add Breez service source (prefer real SDK; fall back to stub for easier installs)
if(BREEZ_SDK_FOUND)
    message(STATUS "Including Breez SDK implementation")
    target_sources(${PLUGIN_NAME} PRIVATE src/breez-service.cpp)
else()
    message(STATUS "Using stub Breez service (Breez SDK not found)")
    target_sources(${PLUGIN_NAME} PRIVATE src/breez-service-stub.cpp)
endif()

target_link_libraries(${PLUGIN_NAME} PRIVATE
    Qt6::Widgets
    Qt6::Svg
    Qt6::Network
    OBS::libobs
    OBS::frontend-api
)

if(BREEZ_SDK_FOUND)
    target_link_libraries(${PLUGIN_NAME} PRIVATE ${BREEZ_SDK_LIBRARIES})
endif()

# Set output directory for the plugin
set_target_properties(${PLUGIN_NAME} PROPERTIES
    PREFIX ""
    LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/$<CONFIG>/$<PLATFORM_ID>/
    RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/$<CONFIG>/$<PLATFORM_ID>/
    OUTPUT_NAME ${PLUGIN_NAME}
)

# Install plugin files
install(TARGETS ${PLUGIN_NAME}
    LIBRARY DESTINATION ${OBS_PLUGIN_DESTINATION}
    RUNTIME DESTINATION ${OBS_PLUGIN_DESTINATION}
)

# Install data files
install(DIRECTORY data/
    DESTINATION ${OBS_DATA_PATH}/${PLUGIN_NAME}
    USE_SOURCE_PERMISSIONS
)

# Add uninstall target
if(NOT TARGET uninstall)
    configure_file(
        "${CMAKE_CURRENT_SOURCE_DIR}/cmake/cmake_uninstall.cmake.in"
        "${CMAKE_CURRENT_BINARY_DIR}/cmake_uninstall.cmake"
        IMMEDIATE @ONLY
    )

    add_custom_target(uninstall
        COMMAND ${CMAKE_COMMAND} -P ${CMAKE_CURRENT_BINARY_DIR}/cmake_uninstall.cmake
    )
endif()
