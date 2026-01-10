import { createTool } from '@mastra/core/tools';
import z from 'zod';
import { readFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const llmsTxtPath = join(__dirname, '../../../llms.txt');

export const documentationSearchTool = createTool({
  id: 'documentation-search',
  description: 'Search the web for Nixopus documentation',
  inputSchema: z.object({
    query: z.string().min(1).max(200).describe('The search query'),
  }),
  outputSchema: z.array(
    z.object({
      title: z.string().nullable(),
      url: z.string(),
      content: z.string(),
    }),
  ),
  execute: async ({ context }) => {
    const { query } = context;
    
    let docStructure = '';
    try {
      docStructure = readFileSync(llmsTxtPath, 'utf-8');
    } catch {
    }
    
    const searchQuery = `site:nixopus.com OR site:github.com/nixopus OR "Nixopus" ${query}`;
    
    const response = await fetch(
      `https://api.duckduckgo.com/?q=${encodeURIComponent(searchQuery)}&format=json&no_html=1&skip_disambig=1`
    );
    
    const data = await response.json();
    
    const results = [];
    
    if (data.RelatedTopics && Array.isArray(data.RelatedTopics)) {
      for (const topic of data.RelatedTopics) {
        if (topic.FirstURL && topic.Text) {
          results.push({
            title: topic.Text.split(' - ')[0] || null,
            url: topic.FirstURL,
            content: topic.Text,
          });
        }
      }
    }
    
    if (data.AbstractText) {
      results.push({
        title: data.Heading || null,
        url: data.AbstractURL || '',
        content: data.AbstractText,
      });
    }
    
    if (docStructure) {
      results.push({
        title: 'Nixopus Documentation',
        url: '',
        content: docStructure.slice(0, 500),
      });
    }
    
    return results;
  },
});

