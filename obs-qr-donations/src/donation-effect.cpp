// DonationEffect removed - keep empty stub implementation so references compile.
#include "donation-effect.hpp"

DonationEffect::DonationEffect(QWidget *parent) : QWidget(parent) {}

void DonationEffect::triggerEffect(double, const QString &) {}

void DonationEffect::paintEvent(QPaintEvent *) {}

void DonationEffect::updateEffect() {}

void DonationEffect::createParticles(int, const QRectF &) {}
