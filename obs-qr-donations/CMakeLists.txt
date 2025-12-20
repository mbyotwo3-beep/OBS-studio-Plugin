cmake_minimum_required(VERSION 3.16)
project(obs-qr-donations VERSION 1.0.0 LANGUAGES CXX)

# Allow passing BREEZ_API_KEY from command line or environment
# Usage: cmake -DBREEZ_API_KEY="your_key" or export BREEZ_API_KEY="your_key"
if(DEFINED ENV{BREEZ_API_KEY})
    set(BREEZ_API_KEY $ENV{BREEZ_API_KEY})
endif()

if(BREEZ_API_KEY)
    add_compile_definitions(BREEZ_DEFAULT_API_KEY="${BREEZ_API_KEY}")
    message(STATUS "Using default Breez API Key provided during build")
else()
    message(STATUS "No default Breez API Key provided - users must configure manually")
endif()

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTORCC ON)
set(CMAKE_AUTOUIC ON)

# Find required packages
find_package(Qt6 COMPONENTS Widgets Svg Network Multimedia Concurrent REQUIRED)

if(UNIX AND NOT APPLE)
    find_package(PkgConfig REQUIRED)
    pkg_check_modules(LIBOBS REQUIRED libobs)
    
    # Define OBS::libobs target
    if(NOT TARGET OBS::libobs)
        add_library(OBS::libobs INTERFACE IMPORTED)
        target_include_directories(OBS::libobs INTERFACE ${LIBOBS_INCLUDE_DIRS})
        target_link_libraries(OBS::libobs INTERFACE ${LIBOBS_LIBRARIES})
        target_link_options(OBS::libobs INTERFACE ${LIBOBS_LDFLAGS})
    endif()

    # Find obs-frontend-api manually
    find_library(OBS_FRONTEND_API_LIB NAMES obs-frontend-api)
    find_path(OBS_FRONTEND_API_INCLUDE_DIR NAMES obs-frontend-api.h PATH_SUFFIXES obs)
    
    if(NOT TARGET OBS::frontend-api)
        add_library(OBS::frontend-api INTERFACE IMPORTED)
        target_include_directories(OBS::frontend-api INTERFACE ${OBS_FRONTEND_API_INCLUDE_DIR})
        target_link_libraries(OBS::frontend-api INTERFACE ${OBS_FRONTEND_API_LIB})
    endif()
else()
    find_package(LibOBS REQUIRED)
    find_package(obs-frontend-api REQUIRED)
endif()

# Breez SDK setup
# Allow forcing the stub or overriding detection for easier local testing/CI
option(BREEZ_USE_STUB "Force using the Breez stub implementation even if the SDK is present" OFF)
option(BREEZ_STUB_SIMULATE "When using the Breez stub, simulate invoice creation and sends for UI testing" ON)
set(BREEZ_SDK_PATH "${CMAKE_CURRENT_SOURCE_DIR}/third_party/breez_sdk" CACHE PATH "Path to Breez SDK")

if(BREEZ_USE_STUB)
    message(WARNING "BREEZ_USE_STUB is ON — forcing use of the Breez stub implementation")
    set(BREEZ_SDK_FOUND FALSE)
else()
    if(EXISTS "${BREEZ_SDK_PATH}/include/breez_sdk/breez_sdk.h")
        set(BREEZ_SDK_FOUND TRUE)
        set(BREEZ_SDK_INCLUDE_DIRS "${BREEZ_SDK_PATH}/include")
        set(BREEZ_SDK_LIBRARIES "${BREEZ_SDK_PATH}/lib/libbreez_sdk.so")
        include_directories(${BREEZ_SDK_INCLUDE_DIRS})
        add_compile_definitions(HAVE_BREEZ_SDK)
        message(STATUS "Breez SDK found at ${BREEZ_SDK_PATH}")
    else()
        message(WARNING "Breez SDK not found at ${BREEZ_SDK_PATH}. Spark wallet features will be disabled.")
        set(BREEZ_SDK_FOUND FALSE)
    endif()
endif()

if(BREEZ_USE_STUB AND BREEZ_STUB_SIMULATE)
    # add_compile_definitions(BREEZ_STUB_SIMULATE) - Moved to target specific
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
    src/backup-reminder-dialog.cpp
    src/send-payment-dialog.cpp
    src/manage-wallet-dialog.cpp
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
    target_sources(${PLUGIN_NAME} PRIVATE third_party/breez_sdk/include/breez_sdk/breez_sdk.cpp)
else()
    message(STATUS "Using stub Breez service (Breez SDK not found)")
    target_sources(${PLUGIN_NAME} PRIVATE src/breez-service-stub.cpp)
endif()

target_link_libraries(${PLUGIN_NAME} PRIVATE
    Qt6::Widgets
    Qt6::Svg
    Qt6::Network
    Qt6::Multimedia
    OBS::libobs
    OBS::frontend-api
    Qt6::Concurrent
)

if(BREEZ_USE_STUB AND BREEZ_STUB_SIMULATE)
    target_compile_definitions(${PLUGIN_NAME} PRIVATE BREEZ_STUB_SIMULATE)
endif()

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

if(NOT DEFINED OBS_PLUGIN_DESTINATION)
    set(OBS_PLUGIN_DESTINATION "${CMAKE_INSTALL_PREFIX}/lib/obs-plugins")
endif()

if(NOT DEFINED OBS_DATA_PATH)
    set(OBS_DATA_PATH "${CMAKE_INSTALL_PREFIX}/share/obs/obs-plugins")
endif()

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
# if(NOT TARGET uninstall)
#     configure_file(
#         "${CMAKE_CURRENT_SOURCE_DIR}/cmake/cmake_uninstall.cmake.in"
#         "${CMAKE_CURRENT_BINARY_DIR}/cmake_uninstall.cmake"
#         IMMEDIATE @ONLY
#     )
# 
#     add_custom_target(uninstall
#         COMMAND ${CMAKE_COMMAND} -P ${CMAKE_CURRENT_BINARY_DIR}/cmake_uninstall.cmake
#     )
# endif()
# Test configuration
enable_testing()
add_subdirectory(tests)
