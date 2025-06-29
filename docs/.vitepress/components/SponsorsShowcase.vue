<template>
  <div v-if="isLoading" class="loading-container">
    <div class="loading-spinner"></div>
    <p>Loading sponsors...</p>
  </div>

  <div v-else-if="error" class="error-container">
    <p>Error loading sponsors: {{ error }}</p>
    <details>
      <summary>Debug Info</summary>
      <pre>{{
        JSON.stringify({ error: error, url: "/sponsor/sponsors.json" }, null, 2)
      }}</pre>
    </details>
  </div>

  <div v-else-if="sponsors.length === 0" class="error-container">
    <p>No sponsors data found</p>
    <details>
      <summary>Debug Info</summary>
      <pre>{{
        JSON.stringify({ sponsors: sponsors, length: sponsors.length }, null, 2)
      }}</pre>
    </details>
  </div>

  <div v-else class="sponsors-showcase">
    <div class="sponsors-header">
      <h3 class="sponsors-title">Meet Our Amazing Sponsors</h3>
      <div class="sponsors-count-badge">
        {{ sponsors.length }} sponsor{{ sponsors.length !== 1 ? "s" : "" }}
      </div>
    </div>

    <div class="sponsors-grid">
      <div v-for="sponsor in sponsors" :key="sponsor.name" class="sponsor-card">
        <div class="card-header">
          <div class="sponsor-avatar">
            <img :src="sponsor.avatar" :alt="sponsor.name" />
          </div>
          <div class="sponsor-basic-info">
            <h4 class="sponsor-name">{{ sponsor.name }}</h4>
            <div class="sponsor-type">
              <svg
                v-if="sponsor.type === 'Individual'"
                width="12"
                height="12"
                fill="currentColor"
                viewBox="0 0 16 16"
              >
                <path
                  d="M8 8a3 3 0 1 0 0-6 3 3 0 0 0 0 6Zm2-3a2 2 0 1 1-4 0 2 2 0 0 1 4 0Zm4 8c0 1-1 1-1 1H3s-1 0-1-1 1-4 6-4 6 3 6 4Zm-1-.004c-.001-.246-.154-.986-.832-1.664C11.516 10.68 10.289 10 8 10c-2.29 0-3.516.68-4.168 1.332-.678.678-.83 1.418-.832 1.664h10Z"
                />
              </svg>
              <svg
                v-else
                width="12"
                height="12"
                fill="currentColor"
                viewBox="0 0 16 16"
              >
                <path
                  d="M2.5 3A1.5 1.5 0 0 0 1 4.5v.793c.026.009.051.02.076.032L7.674 8.51c.206.1.446.1.652 0l6.598-3.185A.755.755 0 0 1 15 5.293V4.5A1.5 1.5 0 0 0 13.5 3h-11Z"
                />
                <path
                  d="M15 6.954 8.978 9.86a2.25 2.25 0 0 1-1.956 0L1 6.954V11.5A1.5 1.5 0 0 0 2.5 13h11a1.5 1.5 0 0 0 1.5-1.5V6.954Z"
                />
              </svg>
              {{ sponsor.type }}
            </div>
          </div>
          <div v-if="sponsor.contribution" class="contribution-badge">
            {{ sponsor.contribution }}
          </div>
        </div>

        <div class="sponsor-description">
          {{ sponsor.description }}
        </div>

        <div class="sponsor-actions">
          <a
            v-if="sponsor.github"
            :href="sponsor.github"
            target="_blank"
            class="action-btn github-btn"
          >
            <svg width="16" height="16" fill="currentColor" viewBox="0 0 16 16">
              <path
                d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.012 8.012 0 0 0 16 8c0-4.42-3.58-8-8-8z"
              />
            </svg>
            GitHub
          </a>

          <a
            v-if="sponsor.website"
            :href="sponsor.website"
            target="_blank"
            class="action-btn portfolio-btn"
          >
            <svg width="16" height="16" fill="currentColor" viewBox="0 0 16 16">
              <path
                d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8zm7.5-6.923c-.67.204-1.335.82-1.887 1.855A7.97 7.97 0 0 0 5.145 4H7.5V1.077zM4.09 4a9.267 9.267 0 0 1 .64-1.539 6.7 6.7 0 0 1 .597-.933A7.025 7.025 0 0 0 2.255 4H4.09zm-.582 3.5c.03-.877.138-1.718.312-2.5H1.674a6.958 6.958 0 0 0-.656 2.5h2.49zM4.847 5a12.5 12.5 0 0 0-.338 2.5H7.5V5H4.847zM8.5 5v2.5h2.99a12.495 12.495 0 0 0-.337-2.5H8.5zM4.51 8.5a12.5 12.5 0 0 0 .337 2.5H7.5V8.5H4.51zm3.99 0V11h2.653c.187-.765.306-1.608.338-2.5H8.5zM5.145 12c.138.386.295.744.468 1.068.552 1.035 1.218 1.65 1.887 1.855V12H5.145zm.182 2.472a6.696 6.696 0 0 1-.597-.933A9.268 9.268 0 0 1 4.09 12H2.255a7.024 7.024 0 0 0 3.072 2.472zM3.82 11a13.652 13.652 0 0 1-.312-2.5h-2.49c.062.89.291 1.733.656 2.5H3.82zm6.853 3.472A7.024 7.024 0 0 0 13.745 12H11.91a9.27 9.27 0 0 1-.64 1.539 6.688 6.688 0 0 1-.597.933zM8.5 12v2.923c.67-.204 1.335-.82 1.887-1.855.173-.324.33-.682.468-1.068H8.5zm3.68-1h2.146c.365-.767.594-1.61.656-2.5h-2.49a13.65 13.65 0 0 1-.312 2.5zm2.802-3.5a6.959 6.959 0 0 0-.656-2.5H12.18c.174.782.282 1.623.312 2.5h2.49zM11.27 2.461c.247.464.462.98.64 1.539h1.835a7.024 7.024 0 0 0-3.072-2.472c.218.284.418.598.597.933zM10.855 4a7.966 7.966 0 0 0-.468-1.068C9.835 1.897 9.17 1.282 8.5 1.077V4h2.355z"
              />
            </svg>
            Portfolio
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";

const sponsors = ref([]);
const isLoading = ref(true);
const error = ref(null);

onMounted(async () => {
  console.log("SponsorsShowcase: Component mounted");

  try {
    console.log("SponsorsShowcase: Starting fetch...");

    // TODO: Issue due to url fetch in local vs Github pages build
    const urls = [
      "/sponsor/sponsors.json",
      "./sponsors.json",
      window.location.origin + "/sponsor/sponsors.json"
    ];

    let response = null;
    let lastError = null;

    for (const url of urls) {
      try {
        console.log("SponsorsShowcase: Trying URL:", url);
        response = await fetch(url);
        if (response.ok) {
          console.log("SponsorsShowcase: Successfully fetched from:", url);
          break;
        } else {
          console.log(
            "SponsorsShowcase: Failed to fetch from:",
            url,
            "Status:",
            response.status
          );
        }
      } catch (e) {
        console.log("SponsorsShowcase: Error fetching from:", url, e.message);
        lastError = e;
      }
    }

    if (!response || !response.ok) {
      throw (
        lastError ||
        new Error(`HTTP ${response?.status}: ${response?.statusText}`)
      );
    }

    const text = await response.text();
    console.log("SponsorsShowcase: Response text length:", text.length);

    const data = JSON.parse(text);
    console.log("SponsorsShowcase: Parsed data:", data);

    sponsors.value = data;
    console.log("SponsorsShowcase: Sponsors data set successfully");
  } catch (err) {
    console.error("SponsorsShowcase: Error loading sponsors:", err);
    error.value = err.message;
  } finally {
    isLoading.value = false;
    console.log(
      "SponsorsShowcase: Loading finished, isLoading:",
      isLoading.value
    );
  }
});
</script>

<style scoped>
/* Loading and Error States */
.loading-container,
.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  color: #9ca3af;
  background: #1a1a1a;
  border: 1px solid #333;
  border-radius: 12px;
  margin: 2rem 0;
}

.loading-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid #333;
  border-top: 3px solid #8b5cf6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

.error-container details {
  margin-top: 1rem;
  color: #6b7280;
  font-size: 0.9rem;
}

.error-container pre {
  white-space: pre-wrap;
  background: #111111;
  color: #d1d5db;
  padding: 1rem;
  border-radius: 8px;
  overflow-x: auto;
  border: 1px solid #2a2a2a;
}

/* Sponsors Showcase */
.sponsors-showcase {
  margin: 2rem 0;
}

.sponsors-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 2rem;
}

.sponsors-title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 800;
  color: #ffffff;
}

.sponsors-count-badge {
  padding: 0.5rem 1rem;
  border-radius: 20px;
  background: linear-gradient(135deg, #8b5cf6, #3b82f6);
  color: white;
  font-size: 0.85rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.sponsors-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.5rem;
}

/* Sponsor Cards -for dark theme */
.sponsor-card {
  background: #1a1a1a;
  border: 1px solid #333;
  border-radius: 12px;
  padding: 1.5rem;
  transition: all 0.3s ease;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  position: relative;
  overflow: hidden;
}

.sponsor-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.4);
  border-color: #555;
}

.sponsor-card::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, #8b5cf6, #3b82f6, #f59e0b);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.sponsor-card:hover::before {
  opacity: 1;
}

.card-header {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 1rem;
}

.sponsor-avatar {
  flex-shrink: 0;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  overflow: hidden;
  background: #2a2a2a;
  border: 2px solid #333;
}

.sponsor-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.sponsor-basic-info {
  flex: 1;
  min-width: 0;
}

.sponsor-name {
  margin: 0 0 0.25rem 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: #ffffff;
  line-height: 1.3;
}

.sponsor-type {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.8rem;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-weight: 500;
}

.sponsor-type svg {
  opacity: 0.7;
}

.contribution-badge {
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  background: linear-gradient(135deg, #10b981, #059669);
  color: white;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  white-space: nowrap;
}

.sponsor-description {
  color: #d1d5db;
  font-size: 0.9rem;
  line-height: 1.5;
  margin-bottom: 1.25rem;
}

.sponsor-actions {
  display: flex;
  gap: 0.75rem;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 500;
  text-decoration: none;
  transition: all 0.2s ease;
  border: 1px solid transparent;
}

.github-btn {
  background: #24292e;
  color: #ffffff;
  border-color: #444;
}

.github-btn:hover {
  background: #2f363d;
  border-color: #666;
  transform: translateY(-1px);
}

.portfolio-btn {
  background: transparent;
  color: #9ca3af;
  border-color: #444;
}

.portfolio-btn:hover {
  background: #374151;
  color: #ffffff;
  border-color: #666;
  transform: translateY(-1px);
}

.action-btn svg {
  flex-shrink: 0;
}

/* Responsive Design */
@media (max-width: 768px) {
  .sponsors-grid {
    grid-template-columns: 1fr;
    gap: 1rem;
  }

  .sponsors-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .sponsor-card {
    padding: 1rem;
  }

  .card-header {
    gap: 0.75rem;
  }

  .sponsor-avatar {
    width: 40px;
    height: 40px;
  }

  .sponsor-actions {
    gap: 0.5rem;
  }

  .action-btn {
    padding: 0.4rem 0.8rem;
    font-size: 0.75rem;
  }
}

@media (max-width: 480px) {
  .sponsors-showcase {
    margin: 1rem 0;
  }

  .sponsors-header {
    margin-bottom: 1.5rem;
  }

  .sponsor-actions {
    flex-direction: column;
  }

  .action-btn {
    justify-content: center;
  }
}
</style>
