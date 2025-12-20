#include "../src/breez-service.hpp"
#include <QSignalSpy>
#include <QTimer>
#include <gtest/gtest.h>

class BreezSendStubTest : public ::testing::Test {
protected:
  BreezService *service = nullptr;

  void SetUp() override { service = &BreezService::instance(); }
};

TEST_F(BreezSendStubTest, SendLightningEmitsCompleted) {
#ifdef HAVE_BREEZ_SDK
  GTEST_SKIP() << "Skipping stub send test - Breez SDK detected";
#endif
  QSignalSpy spy(service, SIGNAL(sendCompleted(bool, QString)));
  ASSERT_TRUE(spy.isValid());

  // Call sendLightningPayment — stub will emit sendCompleted(false, ...)
  service->sendLightningPayment("lnstubtest");

  // Wait for signal (up to 5 seconds)
  bool got = spy.wait(5000);
  EXPECT_TRUE(got) << "sendCompleted signal was not emitted within timeout";
  EXPECT_GT(spy.count(), 0);

  // Check last signal args
  QList<QVariant> args = spy.takeFirst();
  bool ok = args.at(0).toBool();
  QString msg = args.at(1).toString();
  // Stub may simulate success or failure; ensure we received a non-empty
  // message/txid
  EXPECT_FALSE(msg.isEmpty());
}

TEST_F(BreezSendStubTest, SendOnChainEmitsCompleted) {
#ifdef HAVE_BREEZ_SDK
  GTEST_SKIP() << "Skipping stub send test - Breez SDK detected";
#endif
  QSignalSpy spy(service, SIGNAL(sendCompleted(bool, QString)));
  ASSERT_TRUE(spy.isValid());

  // Call sendOnChain — stub will emit sendCompleted(false, ...)
  service->sendOnChain("sampleaddress", 1000, "bitcoin");

  // Wait for signal (up to 5 seconds)
  bool got = spy.wait(5000);
  EXPECT_TRUE(got) << "sendCompleted signal was not emitted within timeout";
  EXPECT_GT(spy.count(), 0);

  // Check last signal args
  QList<QVariant> args = spy.takeFirst();
  bool ok = args.at(0).toBool();
  QString msg = args.at(1).toString();
  EXPECT_FALSE(msg.isEmpty());
}
