# Test configuration for OBS QR Donations Plugin

# Find required packages
find_package(Qt6 REQUIRED COMPONENTS Core Test)
find_package(GTest REQUIRED)

# Add test executable
# Define common sources needed for tests
set(TEST_SOURCES
    ${CMAKE_SOURCE_DIR}/src/breez-service-stub.cpp
    ${CMAKE_SOURCE_DIR}/src/breez-service.hpp
    ${CMAKE_SOURCE_DIR}/src/asset-manager.cpp
)

# Add test executables
add_executable(test_breez_integration
    test_breez_integration.cpp
    test_main.cpp
    ${TEST_SOURCES}
)

add_executable(test_breez_stub
    test_breez_stub.cpp
    test_main.cpp
    ${TEST_SOURCES}
)

add_executable(test_breez_send_stub
    test_breez_send_stub.cpp
    test_main.cpp
    ${TEST_SOURCES}
)

# Configure test targets
foreach(TEST_TARGET test_breez_integration test_breez_stub test_breez_send_stub)
    target_link_libraries(${TEST_TARGET} PRIVATE
        Qt6::Core
        Qt6::Test
        Qt6::Network
        Qt6::Concurrent
        GTest::GTest
        GTest::Main
    )
    
    target_include_directories(${TEST_TARGET} PRIVATE
        ${CMAKE_SOURCE_DIR}/src
        ${GTEST_INCLUDE_DIRS}
    )
    
    # Add definitions needed for compilation
    target_compile_definitions(${TEST_TARGET} PRIVATE
        BREEZ_USE_STUB
    )
    
    set_target_properties(${TEST_TARGET} PROPERTIES
        RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/tests
    )
endforeach()

# Enable simulation only for integration and send stub tests
target_compile_definitions(test_breez_integration PRIVATE BREEZ_STUB_SIMULATE)
target_compile_definitions(test_breez_send_stub PRIVATE BREEZ_STUB_SIMULATE)

add_test(NAME breez_stub_test COMMAND test_breez_stub)
add_test(NAME breez_send_stub_test COMMAND test_breez_send_stub)
add_test(NAME breez_integration_test COMMAND test_breez_integration)

# Copy test resources
file(COPY ${CMAKE_SOURCE_DIR}/resources DESTINATION ${CMAKE_BINARY_DIR}/tests)

# Add custom test target
add_custom_target(run-tests
    COMMAND ${CMAKE_CTEST_COMMAND} --output-on-failure
    WORKING_DIRECTORY ${CMAKE_BINARY_DIR}
    DEPENDS test_breez_integration test_breez_stub test_breez_send_stub
    COMMENT "Running tests..."
)
