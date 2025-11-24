#pragma once

#include <QWidget>
#include <QTimer>
#include <QElapsedTimer>
#include <QPainter>
#include <QPropertyAnimation>
#include <QLabel>

class DonationEffect : public QWidget {
    Q_OBJECT
    Q_PROPERTY(qreal opacity READ opacity WRITE setOpacity)
    
public:
    explicit DonationEffect(QWidget *parent = nullptr);
    ~DonationEffect() override = default;
    
    // Start the effect with the specified amount and currency
    void triggerEffect(double amount, const QString &currency, const QString &memo = "");
    
    // Set effect duration in milliseconds
    void setDuration(int durationMs) { m_durationMs = durationMs; }
    
    // Set effect color
    void setEffectColor(const QColor &color) { m_effectColor = color; }
    
    // Set particle count (more = more intense)
    void setParticleCount(int count) { m_particleCount = count; }
    
    qreal opacity() const { return m_opacity; }
    void setOpacity(qreal opacity);
    
protected:
    void paintEvent(QPaintEvent *event) override;
    void showEvent(QShowEvent *event) override;
    
private slots:
    void updateEffect();
    void onAnimationFinished();
    
private:
    struct Particle {
        QPointF position;
        QPointF velocity;
        qreal size;
        qreal opacity;
        qreal rotation;
        qreal rotationSpeed;
        QColor color;
    };
    
    QTimer m_timer;
    QElapsedTimer m_elapsedTimer;
    QList<Particle> m_particles;
    QColor m_effectColor;
    int m_durationMs;
    int m_particleCount;
    bool m_active;
    qreal m_opacity;
    
    // Notification overlay
    QLabel *m_notificationLabel;
    QPropertyAnimation *m_slideAnimation;
    QPropertyAnimation *m_fadeAnimation;
    
    QString m_donationText;
    QString m_memoText;
    
    void createParticles(int count, const QRectF &area);
    void updateParticles(qreal deltaTime);
    QColor getDonationColor(double amount) const;
};

