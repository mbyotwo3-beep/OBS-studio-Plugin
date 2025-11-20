# Test configuration for OBS QR Donations Plugin

# Find required packages
find_package(Qt6 REQUIRED COMPONENTS Core Test)
find_package(GTest REQUIRED)

# Add test executable
add_executable(test_breez_integration
    test_breez_integration.cpp
)

# Link libraries
target_link_libraries(test_breez_integration PRIVATE
    obs-qr-donations
    Qt6::Core
    GTest::GTest
    GTest::Main
)

# Set include directories
target_include_directories(test_breez_integration PRIVATE
    ${CMAKE_SOURCE_DIR}/src
    ${GTEST_INCLUDE_DIRS}
)

# Add test
add_test(NAME breez_integration_test COMMAND test_breez_integration)

# Set working directory for tests
set_target_properties(test_breez_integration PROPERTIES
    RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/tests
)

# Copy test resources
file(COPY ${CMAKE_SOURCE_DIR}/resources DESTINATION ${CMAKE_BINARY_DIR}/tests)

# Add custom test target
add_custom_target(run-tests
    COMMAND ${CMAKE_CTEST_COMMAND} --output-on-failure
    WORKING_DIRECTORY ${CMAKE_BINARY_DIR}
    DEPENDS test_breez_integration
    COMMENT "Running tests..."
)
