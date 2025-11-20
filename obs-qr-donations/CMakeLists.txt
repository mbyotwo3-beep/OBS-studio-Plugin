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

# Breez SDK setup
set(BREEZ_SDK_PATH "${CMAKE_CURRENT_SOURCE_DIR}/third_party/breez_sdk" CACHE PATH "Path to Breez SDK")

if(EXISTS "${BREEZ_SDK_PATH}")
    add_subdirectory(${BREEZ_SDK_PATH} breez_sdk)
    set(BREEZ_SDK_FOUND TRUE)
    message(STATUS "Breez SDK found at ${BREEZ_SDK_PATH}")
    add_definitions(-DHAVE_BREEZ_SDK)
else()
    message(WARNING "Breez SDK not found at ${BREEZ_SDK_PATH}. Spark wallet features will be disabled.")
    set(BREEZ_SDK_FOUND FALSE)
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

target_link_libraries(${PLUGIN_NAME} PRIVATE
    Qt6::Widgets
    Qt6::Svg
    Qt6::Network
    OBS::libobs
    OBS::frontend-api
    ${BREEZ_SDK_LIBRARIES}
)

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
