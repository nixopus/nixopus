<template>
  <div class="sponsors-marquee-section">
    <div class="container">
      <h2 class="sponsors-title">Our Amazing Sponsors</h2>
      <p class="sponsors-subtitle">
        Thanks to our amazing sponsors who make Nixopus possible
      </p>

      <div v-if="isLoading" class="loading-container">
        <div class="loading-spinner"></div>
        <p>Loading sponsors...</p>
      </div>

      <div v-else-if="error" class="error-container">
        <p>Error loading sponsors: {{ error }}</p>
      </div>

      <div v-else class="sponsors-marquee-container">
        <div class="sponsors-marquee">
          <div class="sponsors-track">
            <div
              v-for="sponsor in duplicatedSponsors"
              :key="`${sponsor.name}-${sponsor.id}`"
              class="sponsor-card-marquee"
              @click="openSponsorLink(sponsor)"
            >
              <div class="sponsor-avatar">
                <img :src="sponsor.avatar" :alt="sponsor.name" />
              </div>
              <div class="sponsor-name">{{ sponsor.name }}</div>
              <div class="sponsor-type">
                {{ sponsor.type }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";

const sponsors = ref([]);
const isLoading = ref(true);
const error = ref(null);

const displaySponsors = computed(() => {
  return sponsors.value || [];
});

const duplicatedSponsors = computed(() => {
  const sponsorsList = displaySponsors.value;
  if (sponsorsList.length === 0) return [];

  // Create multiple copies for infinite scroll effect
  const copies = 3;
  const duplicated = [];

  for (let i = 0; i < copies; i++) {
    sponsorsList.forEach((sponsor, index) => {
      duplicated.push({
        ...sponsor,
        id: `${i}-${index}` // Unique ID for each copy
      });
    });
  }

  return duplicated;
});

const getInitials = (name) => {
  return name
    .replace("@", "")
    .split(" ")
    .map((word) => word.charAt(0).toUpperCase())
    .join("")
    .substring(0, 2);
};

const getTierLabel = (tier) => {
  return "Sponsor";
};

const openSponsorLink = (sponsor) => {
  if (sponsor.website) {
    window.open(sponsor.website, "_blank");
  } else if (sponsor.github) {
    window.open(sponsor.github, "_blank");
  }
};

onMounted(async () => {
  try {
    const response = await fetch("/sponsor/sponsors.json");
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    sponsors.value = await response.json();
  } catch (e) {
    error.value = e.message;
    console.error("Error loading sponsors:", e);
  } finally {
    isLoading.value = false;
  }
});
</script>

<style scoped>
.sponsors-marquee-section {
  padding: 4rem 0;
  background: var(--vp-c-bg);
  overflow: hidden;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
}

.sponsors-title {
  font-size: 2.5rem;
  font-weight: 700;
  color: var(--vp-c-text-1);
  margin: 0 0 1rem 0;
  text-align: center;
}

.sponsors-subtitle {
  font-size: 1.1rem;
  color: var(--vp-c-text-2);
  margin: 0 0 3rem 0;
  line-height: 1.5;
  text-align: center;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  padding: 2rem;
}

.loading-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--vp-c-divider);
  border-top: 3px solid var(--vp-c-brand-1);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

.error-container {
  padding: 2rem;
  color: var(--vp-c-danger-1);
  text-align: center;
}

.sponsors-marquee-container {
  width: 100%;
  overflow: hidden;
  position: relative;
  mask: linear-gradient(90deg, transparent, white 5%, white 95%, transparent);
  -webkit-mask: linear-gradient(
    90deg,
    transparent,
    white 5%,
    white 95%,
    transparent
  );
}

.sponsors-marquee {
  width: 100%;
  position: relative;
}

.sponsors-track {
  display: flex;
  gap: 2rem;
  animation: marquee 40s linear infinite;
  width: max-content;
}

.sponsors-track:hover {
  animation-play-state: paused;
}

@keyframes marquee {
  0% {
    transform: translateX(0);
  }
  100% {
    transform: translateX(-33.333%);
  }
}

.sponsor-card-marquee {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem 1.5rem;
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-divider);
  border-radius: 1rem;
  transition: all 0.3s ease;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  min-width: 180px;
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.sponsor-card-marquee:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
  border-color: var(--vp-c-brand-1);
}

.sponsor-avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  overflow: hidden;
  margin-bottom: 1rem;
  position: relative;
  background: var(--vp-c-gray-light-4);
}

.sponsor-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.sponsor-initials {
  position: absolute;
  top: 1rem;
  right: 1rem;
  width: 32px;
  height: 32px;
  background: var(--vp-c-brand-1);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 600;
}

.sponsor-name {
  font-size: 1rem;
  font-weight: 600;
  color: var(--vp-c-text-1);
  margin-bottom: 0.5rem;
  text-align: center;
}

.sponsor-type {
  font-size: 0.875rem;
  font-weight: 500;
  padding: 0.25rem 0.75rem;
  border-radius: 1rem;
  text-transform: capitalize;
  background: #9be935;
  color: black;
}

/* Dark mode adjustments */
.dark .sponsor-card-marquee {
  background: var(--vp-c-bg-soft);
  border-color: var(--vp-c-divider);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.dark .sponsor-card-marquee:hover {
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.3);
}

/* Responsive design */
@media (max-width: 768px) {
  .sponsors-title {
    font-size: 2rem;
  }

  .sponsors-subtitle {
    font-size: 1rem;
  }

  .container {
    padding: 0 1rem;
  }

  .sponsor-card-marquee {
    min-width: 150px;
    padding: 1.5rem 1rem;
  }

  .sponsors-track {
    gap: 1.5rem;
  }
}

@media (max-width: 480px) {
  .sponsors-marquee-section {
    padding: 3rem 0;
  }

  .sponsor-card-marquee {
    min-width: 130px;
    padding: 1.25rem 0.75rem;
  }

  .sponsors-track {
    gap: 1rem;
  }

  .sponsor-avatar {
    width: 48px;
    height: 48px;
  }

  .sponsor-initials {
    width: 24px;
    height: 24px;
    font-size: 0.65rem;
  }

  .sponsor-name {
    font-size: 0.9rem;
  }

  .sponsor-type {
    font-size: 0.75rem;
    padding: 0.2rem 0.5rem;
  }
}
</style>
