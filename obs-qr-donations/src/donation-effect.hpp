#pragma once

#include <QWidget>
#include <QTimer>
#include <QElapsedTimer>
#include <QPainter>

class DonationEffect : public QWidget {
    Q_OBJECT
    
public:
    explicit DonationEffect(QWidget *parent = nullptr);
    ~DonationEffect() override = default;
    
    // Start the effect with the specified amount and currency
    void triggerEffect(double amount, const QString &currency);
    
    // Set effect duration in milliseconds
    void setDuration(int durationMs) { m_durationMs = durationMs; }
    
    // Set effect color
    void setEffectColor(const QColor &color) { m_effectColor = color; }
    
protected:
    void paintEvent(QPaintEvent *event) override;
    
private slots:
    void updateEffect();
    
private:
    struct Particle {
        QPointF position;
        QPointF velocity;
        qreal size;
        qreal opacity;
    };
    
    QTimer m_timer;
    QElapsedTimer m_elapsedTimer;
    QList<Particle> m_particles;
    QColor m_effectColor;
    int m_durationMs;
    bool m_active;
    
    void createParticles(int count, const QRectF &area);
};
