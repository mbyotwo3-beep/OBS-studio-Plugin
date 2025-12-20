// Single test main that initializes a QCoreApplication for tests that need Qt
// event loop
#include <QCoreApplication>
#include <gtest/gtest.h>

int main(int argc, char **argv) {
  // Initialize Qt application (required for QSignalSpy and other Qt facilities)
  QCoreApplication app(argc, argv);

  ::testing::InitGoogleTest(&argc, argv);
  int result = RUN_ALL_TESTS();

  return result;
}
