#include <gtest/gtest.h>
#include "../src/breez-service.hpp"

class BreezStubTest : public ::testing::Test {
protected:
    void SetUp() override {
        service = &BreezService::instance();
    }

    BreezService *service;
};

TEST_F(BreezStubTest, InitializeWithoutSdk) {
    // When compiled without Breez SDK, initialize should return false. If SDK is present
    // this test is not applicable and should be skipped.
#ifdef HAVE_BREEZ_SDK
    GTEST_SKIP() << "Skipping stub test - Breez SDK detected";
#endif
    QString apiKey = "";
    QString url = "";
    QString key = "";

    bool ok = service->initialize(apiKey, url, key, "bitcoin");
    EXPECT_FALSE(ok);
}
