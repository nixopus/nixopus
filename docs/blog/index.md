---
title: Blog
description: Latest updates and news from the Nixopus team
---

# Blog

Welcome to the Nixopus blog! Here you'll find the latest updates, news, and insights about our project.

<script setup>
import { data as posts } from '../.vitepress/posts.data.mts'
</script>

<div class="blog-list">
  <div v-if="!posts || posts.length === 0" class="no-posts">
    No blog posts found.
  </div>
  <div v-else v-for="post in posts" :key="post.url" class="blog-item">
    <h2>
      <a :href="post.url">{{ post.title }}</a>
    </h2>
    <div class="blog-meta">
      <span class="date">{{ new Date(post.date).toLocaleDateString() }}</span>
      <span class="author">{{ post.author }}</span>
    </div>
    <p class="description">{{ post.description }}</p>
  </div>
</div>

<style>
.blog-list {
  max-width: 800px;
  margin: 0 auto;
}

.blog-item {
  margin-bottom: 2rem;
  padding-bottom: 2rem;
  border-bottom: 1px solid var(--vp-c-divider);
}

.blog-item:last-child {
  border-bottom: none;
}

.blog-meta {
  color: var(--vp-c-text-2);
  font-size: 0.9rem;
  margin: 0.5rem 0;
}

.blog-meta .date {
  margin-right: 1rem;
}

.description {
  color: var(--vp-c-text-1);
  margin-top: 0.5rem;
}

.no-posts {
  text-align: center;
  color: var(--vp-c-text-2);
  padding: 2rem;
}
</style> 