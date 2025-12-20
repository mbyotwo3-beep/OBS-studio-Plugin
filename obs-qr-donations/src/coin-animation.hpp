#pragma once

#include <QElapsedTimer>
#include <QImage>
#include <QQuickPaintedItem>
#include <QTimer>
#include <random>
#include <vector>

class CoinAnimation : public QQuickPaintedItem {
  Q_OBJECT
  Q_PROPERTY(
      int coinCount READ coinCount WRITE setCoinCount NOTIFY coinCountChanged)
  Q_PROPERTY(qreal speed READ speed WRITE setSpeed NOTIFY speedChanged)
  Q_PROPERTY(qreal wind READ wind WRITE setWind NOTIFY windChanged)
  Q_PROPERTY(bool running READ isRunning WRITE setRunning NOTIFY runningChanged)

public:
  explicit CoinAnimation(QQuickItem *parent = nullptr);
  ~CoinAnimation() override = default;

  void paint(QPainter *painter) override;

  // Property getters/setters
  int coinCount() const { return m_coinCount; }
  qreal speed() const { return m_speed; }
  qreal wind() const { return m_wind; }
  bool isRunning() const { return m_running; }

  void setCoinCount(int count);
  void setSpeed(qreal speed);
  void setWind(qreal wind);
  void setRunning(bool running);

  // Start the animation with a specific number of coins
  Q_INVOKABLE void start(int count = 50);

  // Stop the animation
  Q_INVOKABLE void stop();

signals:
  void coinCountChanged();
  void speedChanged();
  void windChanged();
  void runningChanged();
  void finished();

private slots:
  void updateAnimation();

private:
  struct Coin {
    qreal x = 0;
    qreal y = 0;
    qreal velocity = 0;
    qreal rotation = 0;
    qreal rotationSpeed = 0;
    qreal size = 0;
    qreal windEffect = 0;
    QImage image;
  };

  void initializeCoins();
  void resetCoin(Coin &coin);

  std::vector<Coin> m_coins;
  QTimer m_animationTimer;
  QElapsedTimer m_elapsedTimer;
  qint64 m_lastFrameTime = 0;

  int m_coinCount = 50;
  qreal m_speed = 1.0;
  qreal m_wind = 0.0;
  bool m_running = false;

  // Random number generation
  std::mt19937 m_randomGenerator;
  std::uniform_real_distribution<qreal> m_randomDist{0.0, 1.0};

  // Coin images
  QImage m_coinImage;
  QImage m_btcImage;
  QImage m_ethImage;
  QImage m_ltcImage;

  // Animation properties
  static constexpr qreal GRAVITY = 0.005;
  static constexpr qreal MIN_VELOCITY = 0.5;
  static constexpr qreal MAX_VELOCITY = 1.5;
  static constexpr qreal MIN_SIZE = 0.5;
  static constexpr qreal MAX_SIZE = 1.5;
};
