---
layout: home

hero:
  name: "Nixopus Documentation"
  text: ""
  tagline: All the information you need to know about Nixopus
---
<div class="command-block-container">
  <div class="command-block">
    <div class="command-text">
      <span class="prompt">$</span> curl -sSL https://install.nixopus.com | bash
    </div>
    <button class="copy-button" onclick="const text = this.previousElementSibling.textContent.trim().replace(/^\$\s*/, ''); navigator.clipboard.writeText(text); const icon = this.querySelector('.copy-icon'); const text_el = this.querySelector('.copy-text'); icon.innerHTML = '✓'; text_el.innerText='Copied!'; this.classList.add('copied'); setTimeout(()=>{icon.innerHTML = '⧉'; text_el.innerText='Copy'; this.classList.remove('copied');}, 1200);">
      <span class="copy-icon">⧉</span>
      <span class="copy-text">Copy</span>
    </button>
  </div>
</div>
<SponsorsMarquee />
