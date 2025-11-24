#include "donation-effect.hpp"
#include <QPainter>
#include <QRandomGenerator>
#include <QFont>
#include <QFontMetrics>
#include <QVBoxLayout>
#include <cmath>

DonationEffect::DonationEffect(QWidget *parent)
    : QWidget(parent)
    , m_effectColor(QColor(255, 215, 0)) // Gold default
    , m_durationMs(4000) // 4 seconds default
    , m_particleCount(50)
    , m_active(false)
    , m_opacity(1.0)
{
    setAttribute(Qt::WA_TransparentForMouseEvents);
    setAttribute(Qt::WA_TranslucentBackground);
    setWindowFlags(Qt::FramelessWindowHint | Qt::Tool);
    
    // Create notification label
    m_notificationLabel = new QLabel(this);
    m_notificationLabel->setAlignment(Qt::AlignCenter);
    m_notificationLabel->setStyleSheet(
        "QLabel {"
        "   background-color: rgba(76, 175, 80, 220);"
        "   color: white;"
        "   padding: 20px 40px;"
        "   border-radius: 10px;"
        "   font-size: 24px;"
        "   font-weight: bold;"
        "   border: 3px solid rgba(255, 255, 255, 150);"
        "}"
    );
    m_notificationLabel->hide();
    
    // Setup animations
    m_slideAnimation = new QPropertyAnimation(m_notificationLabel, "pos", this);
    m_slideAnimation->setDuration(500);
    m_slideAnimation->setEasingCurve(QEasingCurve::OutCubic);
    
    m_fadeAnimation = new QPropertyAnimation(this, "opacity", this);
    m_fadeAnimation->setDuration(800);
    m_fadeAnimation->setStartValue(1.0);
    m_fadeAnimation->setEndValue(0.0);
    connect(m_fadeAnimation, &QPropertyAnimation::finished, this, &DonationEffect::onAnimationFinished);
    
    // Setup update timer for particles (60 FPS)
    m_timer.setInterval(16);
    connect(&m_timer, &QTimer::timeout, this, &DonationEffect::updateEffect);
}

void DonationEffect::setOpacity(qreal opacity) {
    m_opacity = opacity;
    update();
}

void DonationEffect::triggerEffect(double amount, const QString &currency, const QString &memo) {
    if (m_active) {
        return; // Don't trigger multiple effects simultaneously
    }
    
    m_active = true;
    m_opacity = 1.0;
    
    // Format donation text
    m_donationText = QString("💰 +%1 %2").arg(QString::number(amount, 'f', amount >= 1000 ? 0 : 2)).arg(currency);
    m_memoText = memo;
    
    // Set color based on donation amount
    m_effectColor = getDonationColor(amount);
    
    // Show and position notification
    m_notificationLabel->setText(m_donationText + (!memo.isEmpty() ? QString("\n\"%1\"").arg(memo) : ""));
    m_notificationLabel->adjustSize();
    
    // Start from top, slide down
    QPoint startPos(width() / 2 - m_notificationLabel->width() / 2, -m_notificationLabel->height());
    QPoint endPos(width() / 2 - m_notificationLabel->width() / 2, 50);
    
    m_notificationLabel->move(startPos);
    m_notificationLabel->show();
    
    m_slideAnimation->setStartValue(startPos);
    m_slideAnimation->setEndValue(endPos);
    m_slideAnimation->start();
    
    // Create particles
    createParticles(m_particleCount, rect());
    
    // Start animation
    m_elapsedTimer.start();
    m_timer.start();
    
    // Schedule fade out
    QTimer::singleShot(m_durationMs - 800, this, [this]() {
        m_fadeAnimation->start();
    });
    
    // Auto-hide after duration
    QTimer::singleShot(m_durationMs, this, [this]() {
        if (m_active) {
            m_timer.stop();
            m_active = false;
            hide();
        }
    });
}

void DonationEffect::createParticles(int count, const QRectF &area) {
    m_particles.clear();
    
    for (int i = 0; i < count; ++i) {
        Particle p;
        
        // Start from bottom center, spread outward
        qreal centerX = area.width() / 2;
        qreal spreadX = area.width() * 0.3;
        
        p.position = QPointF(
            centerX + (QRandomGenerator::global()->bounded(2.0) - 1.0) * spreadX,
            area.height() + 10
        );
        
        // Random upward velocity with outward drift
        qreal angle = -M_PI / 2 + (QRandomGenerator::global()->bounded(2.0) - 1.0) * M_PI / 6;
        qreal speed = 100 + QRandomGenerator::global()->bounded(150.0);
        
        p.velocity = QPointF(
            cos(angle) * speed,
            sin(angle) * speed
        );
        
        p.size = 4 + QRandomGenerator::global()->bounded(8.0);
        p.opacity = 1.0;
        p.rotation = QRandomGenerator::global()->bounded(360.0);
        p.rotationSpeed = (QRandomGenerator::global()->bounded(2.0) - 1.0) * 180.0;
        
        // Vary colors around the base effect color
        int hue = m_effectColor.hue();
        int saturation = m_effectColor.saturation();
        int value = m_effectColor.value();
        
        p.color = QColor::fromHsv(
            (hue + QRandomGenerator::global()->bounded(60) - 30 + 360) % 360,
            qMax(0, qMin(255, saturation + QRandomGenerator::global()->bounded(100) - 50)),
            qMax(100, qMin(255, value + QRandomGenerator::global()->bounded(100) - 50))
        );
        
        m_particles.append(p);
    }
}

void DonationEffect::updateEffect() {
    if (!m_active) {
        return;
    }
    
    qreal deltaTime = m_elapsedTimer.restart() / 1000.0; // Convert to seconds
    updateParticles(deltaTime);
    update(); // Trigger repaint
}

void DonationEffect::updateParticles(qreal deltaTime) {
    for (auto &particle : m_particles) {
        // Update position
        particle.position += particle.velocity * deltaTime;
        
        // Apply gravity (slight downward pull)
        particle.velocity.ry() += 50 * deltaTime;
        
        // Air resistance (slow down over time)
        particle.velocity *= 0.98;
        
        // Update rotation
        particle.rotation += particle.rotationSpeed * deltaTime;
        
        // Fade out particles as they rise
        qreal heightRatio = 1.0 - (particle.position.y() / height());
        particle.opacity = qMax(0.0, qMin(1.0, heightRatio * 2.0)) * m_opacity;
    }
}

void DonationEffect::paintEvent(QPaintEvent *event) {
    Q_UNUSED(event);
    
    if (!m_active) {
        return;
    }
    
    QPainter painter(this);
    painter.setRenderHint(QPainter::Antialiasing);
    
    // Draw particles
    for (const auto &particle : m_particles) {
        if (particle.opacity <= 0) {
            continue;
        }
        
        painter.save();
        painter.translate(particle.position);
        painter.rotate(particle.rotation);
        
        QColor color = particle.color;
        color.setAlphaF(particle.opacity);
        painter.setBrush(color);
        painter.setPen(Qt::NoPen);
        
        // Draw different shapes for variety
        int shape = QRandomGenerator::global()->bounded(3);
        if (shape == 0) {
            // Circle
            painter.drawEllipse(QPointF(0, 0), particle.size, particle.size);
        } else if (shape == 1) {
            // Square
            painter.drawRect(-particle.size, -particle.size, particle.size * 2, particle.size * 2);
        } else {
            // Star-like shape
            QPolygonF star;
            for (int i = 0; i < 5; ++i) {
                qreal angle = i * 2 * M_PI / 5 - M_PI / 2;
                star << QPointF(cos(angle) * particle.size, sin(angle) * particle.size);
            }
            painter.drawPolygon(star);
        }
        
        painter.restore();
    }
    
    // Apply overall opacity to notification
    if (m_notificationLabel->isVisible()) {
        painter.setOpacity(m_opacity);
    }
}

void DonationEffect::showEvent(QShowEvent *event) {
    QWidget::showEvent(event);
    // Ensure we're on top and full size of parent
    if (parentWidget()) {
        setGeometry(parentWidget()->rect());
        raise();
    }
}

void DonationEffect::onAnimationFinished() {
    m_notificationLabel->hide();
}

QColor DonationEffect::getDonationColor(double amount) const {
    // Color based on donation size
    if (amount >= 10000) {
        return QColor(148, 0, 211); // Purple - huge donation
    } else if (amount >= 5000) {
        return QColor(255, 0, 255); // Magenta - large donation
    } else if (amount >= 1000) {
        return QColor(255, 215, 0); // Gold - medium donation
    } else if (amount >= 100) {
        return QColor(50, 205, 50); // Lime green - small donation
    } else {
        return QColor(100, 149, 237); // Cornflower blue - tiny donation
    }
}
