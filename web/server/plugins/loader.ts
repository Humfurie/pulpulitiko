export default defineNitroPlugin((nitroApp) => {
  nitroApp.hooks.hook('render:html', (html) => {
    // Inject loader HTML right after <body> tag
    const loaderHtml = `
      <div id="__nuxt_loader">
        <div style="display:flex;flex-direction:column;align-items:center;gap:16px;">
          <div class="pulse" style="font-size:1.875rem;font-weight:800;letter-spacing:-0.025em;">
            <span class="logo-dark">Pulpul</span><span class="logo-accent">itiko</span>
          </div>
          <div class="spinner"></div>
        </div>
      </div>
    `
    html.bodyPrepend.unshift(loaderHtml)
  })
})
