---
layout: home

hero:
  name: "Nixopus"
  text: "Documentation"
  tagline: All the information you need to know about Nixopus
---
<div style="display: flex; justify-content: center; align-items: center; margin: 1.5em 0;">
  <div style="position: relative; display: inline-flex; align-items: center; background: rgba(0,0,0,0.18); border-radius: 12px; border: 1px solid #a78bfa; overflow: hidden; box-shadow: 0 4px 12px rgba(0,0,0,0.15); min-width: 400px; max-width: 90vw;">
        <div style="flex: 1; padding: 1.2em 1.5em; font-family: 'Fira Code', 'Monaco', 'Consolas', monospace; font-size: 1.05em; color: #fff; background: rgba(0,0,0,0.18); overflow-x: auto; white-space: nowrap; max-width: 600px;">
      sudo bash -c "$(curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install-cli.sh)"
    </div>
    <button onclick="navigator.clipboard.writeText('sudo bash -c "$(curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install-cli.sh)"'); this.innerText='Copied!'; setTimeout(()=>this.innerText='Copy',1200);" style="color: #fff; border: none; padding: 1.2em 1.5em; font-weight: 600; font-size: 0.9em; cursor: pointer; transition: all 0.2s; outline: none; border-left: 1px solid #a78bfa; min-width: 80px; display: flex; align-items: center; justify-content: center;">Copy</button>
  </div>
</div>
<SponsorsMarquee />
