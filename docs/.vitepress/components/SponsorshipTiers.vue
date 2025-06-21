<template>
  <div class="sponsorship-tiers">
    <div class="tiers-header">
      <h2>Sponsorship Tiers</h2>
      <p>Choose the sponsorship level that works best for you</p>
    </div>

    <div class="tiers-table">
      <div class="table-header">
        <div class="header-cell tier-col">TIER</div>
        <div class="header-cell contribution-col">CONTRIBUTION</div>
        <div class="header-cell benefits-col">KEY BENEFITS</div>
      </div>

      <div class="table-row partnership-row">
        <div class="tier-cell">
          <div class="tier-info">
            <span class="tier-icon">🤝</span>
            <span class="tier-name">Partnership Sponsor</span>
          </div>
        </div>
        <div class="contribution-cell">
          <span class="contribution-badge partnership"
            >Coupon codes, trials, discounts</span
          >
        </div>
        <div class="benefits-cell">
          Coupon spotlight on homepage & newsletter, Social mentions, "Partners"
          section, Partner-only chat
          <a
            href="mailto:raghavyuva@gmail.com"
            class="tier-action-btn partnership"
            >Contact Us</a
          >
        </div>
      </div>

      <div class="table-row diamond-row">
        <div class="tier-cell">
          <div class="tier-info">
            <span class="tier-icon">💎</span>
            <span class="tier-name">Diamond Sponsor</span>
          </div>
        </div>
        <div class="contribution-cell">
          <span class="contribution-badge diamond">$1,000+ in cash</span>
        </div>
        <div class="benefits-cell">
          Full-suite visibility, Custom integrations, Early-access
          <a
            href="https://github.com/sponsors/raghavyuva"
            target="_blank"
            class="tier-action-btn diamond"
            >Sponsor Now</a
          >
        </div>
      </div>

      <div class="table-row platinum-row">
        <div class="tier-cell">
          <div class="tier-info">
            <span class="tier-icon">🥈</span>
            <span class="tier-name">Platinum Sponsor</span>
          </div>
        </div>
        <div class="contribution-cell">
          <span class="contribution-badge platinum">$500+ in cash</span>
        </div>
        <div class="benefits-cell">
          Logo on docs, Quarterly newsletter, Priority support
          <a
            href="https://github.com/sponsors/raghavyuva"
            target="_blank"
            class="tier-action-btn platinum"
            >Sponsor Now</a
          >
        </div>
      </div>

      <div class="table-row gold-row">
        <div class="tier-cell">
          <div class="tier-info">
            <span class="tier-icon">🥇</span>
            <span class="tier-name">Gold Sponsor</span>
          </div>
        </div>
        <div class="contribution-cell">
          <span class="contribution-badge gold">$100+ in cash</span>
        </div>
        <div class="benefits-cell">
          GitHub badge, Sponsor updates, Sponsor-only chat
          <a
            href="https://github.com/sponsors/raghavyuva"
            target="_blank"
            class="tier-action-btn gold"
            >Sponsor Now</a
          >
        </div>
      </div>

      <div class="table-row silver-row">
        <div class="tier-cell">
          <div class="tier-info">
            <span class="tier-icon">🥉</span>
            <span class="tier-name">Silver Sponsor</span>
          </div>
        </div>
        <div class="contribution-cell">
          <span class="contribution-badge silver">$25+ in cash</span>
        </div>
        <div class="benefits-cell">
          Name listing, Sponsor badge, Updates
          <a
            href="https://github.com/sponsors/raghavyuva"
            target="_blank"
            class="tier-action-btn silver"
            >Sponsor Now</a
          >
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// placeholder only - no script needed for this static component
</script>

<style scoped>
.sponsorship-tiers {
  margin: 3rem 0;
}

.tiers-header {
  text-align: center;
  margin-bottom: 3rem;
}

.tiers-header h2 {
  font-size: 2rem;
  font-weight: 700;
  color: #ffffff;
  margin: 0 0 1rem 0;
}

.tiers-header p {
  font-size: 1rem;
  color: #9ca3af;
  margin: 0;
}

.tiers-table {
  background: #1a1a1a;
  border: 1px solid #333;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
}

.table-header {
  display: grid;
  grid-template-columns: 1fr 1fr 2fr;
  background: #111111;
  border-bottom: 1px solid #333;
}

.header-cell {
  padding: 1rem 1.5rem;
  font-size: 0.65rem;
  font-weight: 600;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 1px;
  border-right: 1px solid #333;
}

.header-cell:last-child {
  border-right: none;
}

.table-row {
  display: grid;
  grid-template-columns: 1fr 1fr 2fr;
  border-bottom: 1px solid #2a2a2a;
  transition: all 0.2s ease;
}

.table-row:last-child {
  border-bottom: none;
}

.table-row:hover {
  background: #242424;
}

.tier-cell,
.contribution-cell,
.benefits-cell {
  padding: 1.5rem;
  border-right: 1px solid #2a2a2a;
  display: flex;
  align-items: center;
}

.tier-cell:last-child,
.contribution-cell:last-child,
.benefits-cell:last-child {
  border-right: none;
}

.tier-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.tier-icon {
  font-size: 1.25rem;
}

.tier-name {
  font-size: 0.7rem !important;
  font-weight: 600;
  color: #ffffff;
}

.contribution-badge {
  padding: 0.375rem 0.75rem;
  border-radius: 16px;
  font-size: 0.5rem !important;
  font-weight: 600;
  color: white;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.contribution-badge.partnership {
  background: linear-gradient(135deg, #10b981, #059669);
}

.contribution-badge.diamond {
  background: linear-gradient(135deg, #8b5cf6, #7c3aed);
}

.contribution-badge.platinum {
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
}

.contribution-badge.gold {
  background: linear-gradient(135deg, #f59e0b, #d97706);
}

.contribution-badge.silver {
  background: linear-gradient(135deg, #6b7280, #4b5563);
}

.benefits-cell {
  font-size: 0.8rem;
  color: #d1d5db;
  line-height: 1.5;
  flex-direction: column;
  align-items: flex-start;
  gap: 1rem;
}

.tier-action-btn {
  display: inline-flex;
  align-items: center;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-size: 0.5rem;
  font-weight: 600;
  text-decoration: none;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  transition: all 0.2s ease;
  margin-top: 0.1rem;
}

.tier-action-btn.partnership {
  background: linear-gradient(135deg, #10b981, #059669);
  color: white;
}

.tier-action-btn.diamond {
  background: linear-gradient(135deg, #8b5cf6, #7c3aed);
  color: white;
}

.tier-action-btn.platinum {
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  color: white;
}

.tier-action-btn.gold {
  background: linear-gradient(135deg, #f59e0b, #d97706);
  color: white;
}

.tier-action-btn.silver {
  background: linear-gradient(135deg, #6b7280, #4b5563);
  color: white;
}

.tier-action-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

/* Responsive Design */
@media (max-width: 1024px) {
  .table-header,
  .table-row {
    grid-template-columns: 1fr;
  }

  .header-cell,
  .tier-cell,
  .contribution-cell,
  .benefits-cell {
    border-right: none;
    border-bottom: 1px solid #2a2a2a;
  }

  .header-cell:last-child,
  .tier-cell:last-child,
  .contribution-cell:last-child,
  .benefits-cell:last-child {
    border-bottom: none;
  }

  .table-row {
    border-bottom: 2px solid #333;
  }

  .table-row:last-child {
    border-bottom: none;
  }
}

@media (max-width: 768px) {
  .tiers-header h2 {
    font-size: 1.75rem;
  }

  .tiers-header p {
    font-size: 0.9rem;
  }

  .header-cell,
  .tier-cell,
  .contribution-cell,
  .benefits-cell {
    padding: 1rem;
  }

  .tier-name {
    font-size: 0.8rem;
  }

  .contribution-badge {
    font-size: 0.65rem;
    padding: 0.3rem 0.6rem;
  }

  .benefits-cell {
    font-size: 0.75rem;
  }
}

@media (max-width: 480px) {
  .sponsorship-tiers {
    margin: 2rem 0;
  }

  .tiers-header {
    margin-bottom: 2rem;
  }

  .tiers-header h2 {
    font-size: 1.5rem;
  }

  .header-cell,
  .tier-cell,
  .contribution-cell,
  .benefits-cell {
    padding: 0.75rem;
  }

  .tier-info {
    gap: 0.5rem;
  }

  .tier-icon {
    font-size: 1rem;
  }

  .tier-name {
    font-size: 0.75rem;
  }
}
</style>
