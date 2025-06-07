# Contributing to Nixopus Documentation

This guide provides detailed instructions for contributing to the Nixopus documentation.

## Documentation Structure

The Nixopus documentation is organized as follows:

```
docs/
├── index.md                   # Documentation homepage
├── architecture/              # System architecture docs
├── code-of-conduct/           # Community guidelines
├── contributing/              # Contribution guides (this section)
├── install/                   # Installation instructions
├── introduction/              # Project introduction
├── self-host/                 # Self-hosting guides
├── file-manager/              # File manager documentation
├── terminal/                  # Terminal documentation
├── preferences/               # User preferences docs
├── operations/                # API operations
├── src/                       # Documentation source files
└── public/                    # Public assets
```

## Setting Up Documentation Development Environment

1. **Prerequisites**
   - Node.js 18.0 or higher
   - Yarn package manager

2. **Environment Setup**
   ```bash
   # Clone the repository
   git clone https://github.com/raghavyuva/nixopus.git
   cd nixopus
   
   # Install documentation dependencies
   cd docs
   yarn install
   ```

3. **Running the Documentation Locally**
   ```bash
   cd docs
   yarn dev
   ```
   
   This will start a local server, typically at [http://localhost:3001](http://localhost:3001).

## Documentation Standards

### Markdown Formatting

Nixopus documentation uses Markdown with some extended features:

1. **Headers**
   - Use `#` for main titles
   - Use `##` for sections
   - Use `###` for subsections
   - Use `####` for subsubsections
   - Don't skip header levels

2. **Code Blocks**
   - Use triple backticks with language for syntax highlighting
   ```markdown
   ```bash
   npm install
   ```
   ```

3. **Links**
   - Internal links: `[Link text](/path/to/page)`
   - External links: `[Link text](https://example.com)`

4. **Images**
   - Store images in `/docs/public/`
   - Reference them: `![Alt text](/image.png)`

5. **Tables**
   ```markdown
   | Header 1 | Header 2 |
   |----------|----------|
   | Cell 1   | Cell 2   |
   ```

6. **Admonitions**
   ```markdown
   > **Note**
   > This is a note.
   
   > **Warning**
   > This is a warning.
   ```

### Content Guidelines

1. **Clarity and Conciseness**
   - Use clear language
   - Keep paragraphs focused on a single topic
   - Avoid jargon without explanation

2. **Structure**
   - Include a table of contents for longer documents
   - Use headers to organize content
   - Use lists for sequential steps or related items

3. **Examples**
   - Provide practical examples
   - Include code snippets where applicable
   - Explain what the examples demonstrate

4. **Screenshots and Diagrams**
   - Use screenshots to illustrate UI elements
   - Use diagrams to explain complex processes
   - Ensure all images have descriptive alt text

## Adding New Documentation

### Creating a New Page

1. **Create a new markdown file**
   - For a new section: `docs/your-section/index.md`
   - For a new topic in an existing section: `docs/existing-section/your-topic.md`

2. **Add frontmatter**
   ```markdown
   ---
   title: "Your Page Title"
   description: "A brief description of your page"
   ---
   ```

3. **Write content following the documentation standards**

4. **Add to navigation**
   - Find the appropriate navigation section and add your page

### Example New Page

```markdown
---
title: "Feature X Guide"
description: "How to effectively use Feature X in Nixopus"
---

# Feature X Guide

This guide explains how to use Feature X in Nixopus.

## Overview

Feature X allows you to [brief description].

## Prerequisites

Before using Feature X, ensure:
- You have [prerequisite 1]
- You have [prerequisite 2]

## Steps to Use Feature X

1. Navigate to the Feature X section in the dashboard.
2. Configure the following settings:
   - Setting 1: [explanation]
   - Setting 2: [explanation]
3. Click "Save" to apply your configuration.

## Advanced Usage

### Scenario 1: [Specific Use Case]

For [specific use case], follow these steps:

```bash
# Example command or code
nixopus feature-x --advanced-option
```

### Scenario 2: [Another Use Case]

[Explanation and steps]

## Troubleshooting

### Common Issue 1

If you encounter [issue], try:
- Solution 1
- Solution 2

### Common Issue 2

[Description and solutions]

## Related Features

- [Related Feature 1](/path/to/related-feature)
- [Related Feature 2](/path/to/another-feature)
```

## Updating Existing Documentation

1. **Identify Issues**
   - Outdated information
   - Missing details
   - Unclear explanations
   - Broken links or references

2. **Make Updates**
   - Use accurate and up-to-date information
   - Keep the existing tone and style consistent
   - Preserve the structure unless it needs improvement
   - Update screenshots if the UI has changed

## Adding API Documentation

Nixopus uses OpenAPI specification for API documentation:

1. **Understanding the Structure**
   - OpenAPI spec is in `docs/src/openapi.json`
   - API routes are documented in the operations directory

2. **Adding a New Endpoint**
   - Update the OpenAPI spec with your new endpoint
   - Add a corresponding Markdown file: `docs/operations/[operationId].md`
   - Update path definitions: `docs/operations/[operationId].paths.mts`

3. **API Documentation Example**
   ```markdown
   ---
   title: "Create Resource"
   description: "API endpoint to create a new resource"
   ---
   
   # Create Resource
   
   This endpoint creates a new resource in the system.
   
   ## Request
   
   `POST /api/resources`
   
   ### Body Parameters
   
   | Parameter | Type   | Required | Description            |
   |-----------|--------|----------|------------------------|
   | name      | string | Yes      | Name of the resource   |
   | type      | string | Yes      | Type of the resource   |
   | config    | object | No       | Additional configuration|
   
   ## Response
   
   ### Success (200 OK)
   
   ```json
   {
     "id": "resource-123",
     "name": "Example Resource",
     "type": "sample",
     "created_at": "2025-06-07T12:00:00Z"
   }
   ```
   
   ### Error Responses
   
   #### 400 Bad Request
   
   ```json
   {
     "error": "Invalid resource parameters"
   }
   ```
   
   #### 401 Unauthorized
   
   ```json
   {
     "error": "Authentication required"
   }
   ```
   ```

## Adding Architecture Documentation

When documenting system architecture:

1. **Include Context**
   - Explain where the component fits in the overall system
   - Describe interactions with other components

2. **Use Diagrams**
   - Component diagrams showing relationships
   - Sequence diagrams for complex processes
   - Data flow diagrams for data-intensive features

3. **Technical Details**
   - Describe implementation details
   - Explain design decisions and trade-offs
   - Include configuration options

## Best Practices for Documentation

1. **Keep It Updated**
   - Review documentation when code changes
   - Mark outdated sections that need updating
   - Remove documentation for deprecated features

2. **Think About the User**
   - Consider the audience's technical level
   - Provide context before diving into details
   - Answer "why" not just "how"

3. **Make It Searchable**
   - Use descriptive headings
   - Include relevant keywords
   - Cross-link related documentation

4. **Get Feedback**
   - Ask other contributors to review
   - Consider feedback from users
   - Test procedures yourself before documenting

## Testing Documentation

1. **Content Review**
   - Verify technical accuracy
   - Check spelling and grammar
   - Ensure links work

2. **Procedure Testing**
   - Follow documented steps to verify they work
   - Test in different environments if applicable
   - Ensure all prerequisites are clearly listed

## Submitting Documentation Changes

1. **Create a Branch**
   ```bash
   git checkout -b docs/your-documentation-change
   ```

2. **Make Your Changes**
   - Follow the guidelines above
   - Test your documentation locally

3. **Commit Your Changes**
   ```bash
   git add .
   git commit -m "docs: add/update documentation for X"
   ```

4. **Submit a Pull Request**
   ```bash
   git push origin docs/your-documentation-change
   ```

5. **Respond to Feedback**
   - Address review comments
   - Make requested changes
   - Explain your approach when needed

## Need Help?

If you need assistance with documentation:
- Ask in the #documentation channel on Discord
- Create an issue for documentation improvements
- Reach out to the documentation maintainers directly

Thank you for improving the Nixopus documentation!
