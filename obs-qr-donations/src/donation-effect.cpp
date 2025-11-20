#include "donation-effect.hpp"
#include <QRandomGenerator>
#include <QPainter>

DonationEffect::DonationEffect(QWidget *parent)
    : QWidget(parent)
    , m_effectColor(255, 215, 0) // Gold color
    , m_durationMs(2000) // 2 seconds
    , m_active(false)
{
    setAttribute(Qt::WA_TransparentForMouseEvents);
    setAttribute(Qt::WA_TranslucentBackground);
    setWindowFlags(Qt::FramelessWindowHint | Qt::Tool | Qt::WindowStaysOnTopHint);
    
    connect(&m_timer, &QTimer::timeout, this, &DonationEffect::updateEffect);
    m_timer.setInterval(16); // ~60 FPS
}

void DonationEffect::triggerEffect(double amount, const QString &currency) {
    if (m_active) {
        return; // Don't interrupt current effect
    }
    
    m_active = true;
    m_particles.clear();
    
    // Create particles based on amount
    int particleCount = qBound(10, static_cast<int>(amount * 2), 100);
    createParticles(particleCount, rect());
    
    m_elapsedTimer.start();
    m_timer.start();
    show();
    raise();
}

void DonationEffect::paintEvent(QPaintEvent *event) {
    if (!m_active || m_particles.isEmpty()) {
        return;
    }
    
    QPainter painter(this);
    painter.setRenderHint(QPainter::Antialiasing);
    
    // Draw particles
    for (const auto &particle : m_particles) {
        if (particle.opacity <= 0.0) {
            continue;
        }
        
        QColor color = m_effectColor;
        color.setAlphaF(particle.opacity);
        
        painter.setPen(Qt::NoPen);
        painter.setBrush(color);
        
        // Draw a simple circle for each particle
        painter.drawEllipse(particle.position, particle.size, particle.size);
    }
}

void DonationEffect::updateEffect() {
    qint64 elapsed = m_elapsedTimer.elapsed();
    qreal progress = qMin(1.0, static_cast<qreal>(elapsed) / m_durationMs);
    
    // Update particles
    for (auto &particle : m_particles) {
        // Apply gravity
        particle.velocity.setY(particle.velocity.y() + 0.1);
        
        // Update position
        particle.position += particle.velocity;
        
        // Fade out
        particle.opacity = 1.0 - progress;
        
        // Slow down
        particle.velocity *= 0.98;
    }
    
    update();
    
    // End effect if duration has passed
    if (elapsed >= m_durationMs) {
        m_timer.stop();
        m_active = false;
        hide();
    }
}

void DonationEffect::createParticles(int count, const QRectF &area) {
    auto *rng = QRandomGenerator::global();
    
    for (int i = 0; i < count; ++i) {
        Particle p;
        
        // Position at a random point in the top half
        p.position = QPointF(
            area.left() + rng->bounded(area.width()),
            area.top() + rng->bounded(area.height() * 0.5)
        );
        
        // Random velocity (up and outwards)
        p.velocity = QPointF(
            (rng->bounded(200) - 100) / 10.0, // -10 to 10
            -rng->bounded(150) / 10.0 - 5     // -5 to -20 (upwards)
        );
        
        // Random size
        p.size = rng->bounded(5) + 2; // 2-7 pixels
        
        // Full opacity initially
        p.opacity = 1.0;
        
        m_particles.append(p);
    }
}
