import { createContentLoader } from 'vitepress'

interface Post {
  title: string
  description: string
  date: string
  author: string
  url: string
}

// Read through all the blog posts and return them as an array of objects
export default createContentLoader('blog/posts/*.md', {
  includeSrc: true,
  transform(rawData) {
    return rawData
      .map(({ frontmatter, url }) => {
        // Skip files without frontmatter
        if (!frontmatter || !frontmatter.title) {
          return null
        }
        return {
          title: frontmatter.title,
          description: frontmatter.description || '',
          date: frontmatter.date || new Date().toISOString(),
          author: frontmatter.author || 'Anonymous',
          url: url.replace(/\.md$/, '')
        }
      })
      .filter((post): post is Post => post !== null)
      .sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime())
  }
}) 