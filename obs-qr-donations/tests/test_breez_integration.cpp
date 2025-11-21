#include <gtest/gtest.h>
#include "../src/breez-service.hpp"
#include <QCoreApplication>
#include <QEventLoop>
#include <QTimer>

class BreezServiceTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Use the shared QCoreApplication provided by tests/test_main.cpp
        // Initialize with test configuration
        QString apiKey = qEnvironmentVariable("BREEZ_API_KEY");
        QString sparkUrl = qEnvironmentVariable("SPARK_URL");
        QString sparkKey = qEnvironmentVariable("SPARK_ACCESS_KEY");

        if (apiKey.isEmpty() || sparkUrl.isEmpty()) {
            GTEST_SKIP() << "Skipping live tests - missing required environment variables";
        }

        service = &BreezService::instance();
        bool initialized = service->initialize(apiKey, sparkUrl, sparkKey, "bitcoin");
        if (!initialized) {
            GTEST_SKIP() << "Failed to initialize Breez service";
        }
    }
    BreezService* service;
};

TEST_F(BreezServiceTest, TestInvoiceCreation) {
    // Test creating a new invoice
    QString invoice = service->createInvoice(1000, "Test invoice", 3600);
    EXPECT_FALSE(invoice.isEmpty()) << "Failed to create invoice";
    
    // Verify invoice format (starts with lnbc for Lightning invoices)
    EXPECT_TRUE(invoice.startsWith("lnbc")) << "Invalid invoice format";
}

TEST_F(BreezServiceTest, TestNodeInfo) {
    QString nodeInfo = service->nodeInfo();
    EXPECT_FALSE(nodeInfo.isEmpty()) << "Failed to get node info";
    
    // Verify node info contains expected fields
    EXPECT_TRUE(nodeInfo.contains("alias")) << "Node info missing alias";
    EXPECT_TRUE(nodeInfo.contains("pubkey")) << "Node info missing pubkey";
}

TEST_F(BreezServiceTest, TestBalance) {
    qint64 balance = service->balance();
    EXPECT_GE(balance, 0) << "Invalid balance";
}

// Add more test cases as needed

// test_main.cpp provides main() and QCoreApplication
