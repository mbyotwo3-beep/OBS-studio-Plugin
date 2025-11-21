# Test configuration for OBS QR Donations Plugin

# Find required packages
find_package(Qt6 REQUIRED COMPONENTS Core Test)
find_package(GTest REQUIRED)

# Add test executable
add_executable(test_breez_integration
    test_breez_integration.cpp
    test_main.cpp
)

add_executable(test_breez_stub
    test_breez_stub.cpp
    test_main.cpp
)

add_executable(test_breez_send_stub
    test_breez_send_stub.cpp
    test_main.cpp
)

target_link_libraries(test_breez_stub PRIVATE
    obs-qr-donations
    Qt6::Core
    Qt6::Test
    GTest::GTest
    GTest::Main
)

target_link_libraries(test_breez_send_stub PRIVATE
    obs-qr-donations
    Qt6::Core
    Qt6::Test
    GTest::GTest
    GTest::Main
)

target_include_directories(test_breez_send_stub PRIVATE
    ${CMAKE_SOURCE_DIR}/src
    ${GTEST_INCLUDE_DIRS}
)

add_test(NAME breez_stub_test COMMAND test_breez_stub)
add_test(NAME breez_send_stub_test COMMAND test_breez_send_stub)

# Link libraries
target_link_libraries(test_breez_integration PRIVATE
    obs-qr-donations
    Qt6::Core
    Qt6::Test
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

set_target_properties(test_breez_send_stub PROPERTIES
    RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/tests
)

# Copy test resources
file(COPY ${CMAKE_SOURCE_DIR}/resources DESTINATION ${CMAKE_BINARY_DIR}/tests)

# Add custom test target
add_custom_target(run-tests
    COMMAND ${CMAKE_CTEST_COMMAND} --output-on-failure
    WORKING_DIRECTORY ${CMAKE_BINARY_DIR}
    DEPENDS test_breez_integration test_breez_stub test_breez_send_stub
    COMMENT "Running tests..."
)
