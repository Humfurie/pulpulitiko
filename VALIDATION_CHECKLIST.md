# Manual Validation Checklist

This document provides step-by-step validation procedures for the security, SEO, accessibility, and performance improvements implemented in Weeks 1-3.

## Prerequisites

Before starting validation, ensure:
- [ ] Application is running (dev, staging, or production)
- [ ] You have access to browser developer tools
- [ ] At least 3 published articles exist for testing

**For local development:**
```bash
# Option 1: Docker stack
docker compose -f docker-compose.prod.yml up -d
# Access at: https://pulpulitiko.humfurie.org

# Option 2: Dev server
cd web && npm run dev
# Access at: http://localhost:3000
```

---

## 1. Schema.org Validation

### 1.1 Google Rich Results Test

**Goal:** Verify NewsArticle and BreadcrumbList structured data is valid

**Steps:**
1. Navigate to any article page on your site
2. Copy the full URL
3. Go to [Google Rich Results Test](https://search.google.com/test/rich-results)
4. Paste the URL and click "Test URL"
5. Wait for results

**Success Criteria:**
- [ ] "Page is eligible for rich results" message appears
- [ ] NewsArticle type detected
- [ ] BreadcrumbList type detected
- [ ] No errors or warnings
- [ ] All required properties present:
  - headline
  - datePublished
  - author (with @type: Person)
  - publisher (with @type: Organization and logo)
  - wordCount (number > 0)
  - articleSection (category name)
  - spatialCoverage (Philippines)
  - isAccessibleForFree (true)

**Repeat for at least 3 different articles**

---

### 1.2 Schema.org Validator

**Goal:** Validate JSON-LD syntax is correct

**Steps:**
1. Navigate to any article page
2. Right-click → "View Page Source"
3. Find and copy the JSON-LD script content (between `<script type="application/ld+json">` tags)
4. Go to [Schema.org Validator](https://validator.schema.org/)
5. Paste the JSON-LD and validate

**Success Criteria:**
- [ ] No validation errors
- [ ] All properties recognized by Schema.org
- [ ] No "undefined" or "null" values in output
- [ ] articleBody is plain text (no HTML tags)
- [ ] datePublished ≤ dateModified (if both present)

---

### 1.3 Google Search Console Sitemap

**Goal:** Ensure search engines can discover all articles

**Steps:**
1. Go to [Google Search Console](https://search.google.com/search-console)
2. Select your property (pulpulitiko.humfurie.org)
3. Navigate to Sitemaps section
4. Submit sitemap URL: `https://pulpulitiko.humfurie.org/sitemap.xml`
5. Wait 24-48 hours

**Success Criteria:**
- [ ] Sitemap successfully submitted
- [ ] No errors in sitemap
- [ ] All article URLs indexed (check after 48 hours)

---

## 2. SEO Meta Tags Validation

### 2.1 Meta Tags Inspection

**Goal:** Verify all required meta tags are present

**Steps:**
1. Navigate to any article page
2. Open browser DevTools (F12)
3. Go to Elements/Inspector tab
4. Inspect `<head>` section

**Check for these tags:**
- [ ] `<title>` - Present and descriptive
- [ ] `<meta name="description">` - Present (< 160 chars recommended)
- [ ] `<meta property="og:title">` - Present
- [ ] `<meta property="og:type" content="article">` - Correct type
- [ ] `<meta property="og:url">` - Present and correct
- [ ] `<meta property="og:image">` - Present (if article has image)
- [ ] `<meta property="og:locale" content="en_US">` - Present
- [ ] `<meta property="article:section">` - Present (category name)
- [ ] `<meta name="robots" content="index, follow">` - Allows indexing
- [ ] `<link rel="canonical">` - Present and matches current URL
- [ ] `<meta name="theme-color">` - Present

---

### 2.2 robots.txt Validation

**Goal:** Verify dynamic robots.txt works correctly

**Steps:**
1. Navigate to `/robots.txt` on your site
2. Inspect content

**Success Criteria:**
- [ ] Returns HTTP 200 status
- [ ] Content-Type: `text/plain; charset=utf-8`
- [ ] Contains `User-agent: *`
- [ ] Contains `Allow: /`
- [ ] Contains `Sitemap:` with correct site URL (not localhost in production)
- [ ] Disallow rules include: `/admin/`, `/api/`, `/account/`, `/login/`, `/register/`
- [ ] Sitemap URL matches current environment

**Test in different environments:**
- [ ] Development (localhost:3000)
- [ ] Staging (if applicable)
- [ ] Production (pulpulitiko.humfurie.org)

---

## 3. Accessibility Validation

### 3.1 WAVE Browser Extension

**Goal:** Identify accessibility violations

**Steps:**
1. Install [WAVE Extension](https://wave.webaim.org/extension/) for Chrome/Firefox
2. Navigate to homepage
3. Click WAVE icon in toolbar
4. Review results

**Test on these pages:**
- [ ] Homepage - Zero errors
- [ ] Article list page - Zero errors
- [ ] Article detail page - Zero errors
- [ ] Admin article editor - Zero errors (if accessible)

**Success Criteria:**
- [ ] Zero errors on all pages
- [ ] No missing alt text on images
- [ ] No missing form labels
- [ ] Proper heading structure (no skipped levels)
- [ ] Sufficient color contrast (WCAG AA)

---

### 3.2 Keyboard Navigation

**Goal:** Verify site is fully keyboard accessible

**Steps:**
1. Navigate to homepage
2. Use only keyboard (no mouse):
   - `Tab` - Move to next interactive element
   - `Shift+Tab` - Move to previous element
   - `Enter/Space` - Activate links/buttons
   - `Esc` - Close modals/dropdowns

**Success Criteria:**
- [ ] All interactive elements are reachable via Tab
- [ ] Focus indicator is clearly visible on all elements
- [ ] Skip-to-content link appears on first Tab press
- [ ] Tab order is logical (top to bottom, left to right)
- [ ] No keyboard traps (can always Tab away)
- [ ] Dropdowns/modals close with Escape key

---

### 3.3 Screen Reader Testing (Optional but Recommended)

**Goal:** Verify content is understandable to screen reader users

**Tools:**
- Windows: [NVDA](https://www.nvaccess.org/) (free)
- macOS: VoiceOver (built-in, Cmd+F5)
- Linux: Orca (built-in on most distros)

**Test these elements:**
- [ ] Word count badge announces correctly: "Article has X words, status: Good"
- [ ] Navigation landmarks are announced (header, main, footer)
- [ ] Breadcrumbs announce correctly
- [ ] Article title, author, and date are clear
- [ ] Images have meaningful alt text

---

### 3.4 Automated Accessibility Tests

**Goal:** Run axe accessibility tests via E2E suite

**Steps:**
```bash
cd web
npm run dev  # Terminal 1
npm run test:e2e -- accessibility.spec.ts  # Terminal 2
```

**Success Criteria:**
- [ ] All 12 accessibility tests pass
- [ ] No WCAG 2.1 AA violations detected

---

## 4. Security Validation

### 4.1 XSS Prevention Testing

**Goal:** Verify HTML sanitization prevents XSS attacks

**⚠️ WARNING:** Only test XSS on your own application with permission

**Backend Test (if you have admin access):**
1. Login to admin panel
2. Create new article
3. In content editor, try inserting:
   ```html
   <script>alert('XSS')</script>
   <img src=x onerror=alert('XSS')>
   <a href="javascript:alert('XSS')">Click me</a>
   ```
4. Save article
5. View article on frontend

**Success Criteria:**
- [ ] No alert boxes execute
- [ ] Script tags are removed from HTML source
- [ ] Event handlers (onerror, onclick) are removed
- [ ] javascript: URLs are removed or neutered
- [ ] Safe HTML (headings, bold, links) is preserved

**Comment Test (if comments are enabled):**
1. Try posting a comment with XSS payload
2. Verify script doesn't execute on page

---

### 4.2 CSP Headers Validation

**Goal:** Verify Content Security Policy headers are present

**Steps:**
1. Navigate to homepage
2. Open DevTools (F12)
3. Go to Network tab
4. Refresh page
5. Click on main document request
6. Go to Headers tab
7. Look for Response Headers

**Success Criteria:**
- [ ] `Content-Security-Policy` or `Content-Security-Policy-Report-Only` header present
- [ ] CSP includes these directives:
  - `default-src 'self'`
  - `img-src 'self' https://minio.humfurie.org https:`
  - `script-src 'self' 'unsafe-inline'`
  - `style-src 'self' 'unsafe-inline'`
  - `connect-src 'self' https://minio.humfurie.org`
- [ ] `X-Content-Type-Options: nosniff` header present
- [ ] `X-Frame-Options` header present

**Monitor CSP Violations:**
1. Check browser console for CSP violation warnings
2. If violations appear, verify they're logged to `/api/csp-report` endpoint

---

### 4.3 E2E XSS Prevention Tests

**Goal:** Run automated XSS prevention tests

**Steps:**
```bash
cd web
npm run dev  # Terminal 1
npm run test:e2e -- xss-prevention.spec.ts  # Terminal 2
```

**Success Criteria:**
- [ ] All 11 XSS prevention tests pass
- [ ] No script tags in article content
- [ ] No inline event handlers
- [ ] No malicious URLs (javascript:, data:, vbscript:)

---

### 4.4 OWASP ZAP Scan (Optional)

**Goal:** Run automated security scan

**Steps:**
1. Download [OWASP ZAP](https://www.zaproxy.org/download/)
2. Launch ZAP
3. Enter your site URL
4. Run automated scan
5. Review results

**Success Criteria:**
- [ ] No high-severity XSS vulnerabilities
- [ ] No SQL injection vulnerabilities
- [ ] Security headers present (CSP, X-Content-Type-Options, etc.)

---

## 5. Performance Validation

### 5.1 Lighthouse Audit

**Goal:** Verify performance, accessibility, and SEO scores

**Steps:**
1. Open site in Chrome
2. Open DevTools (F12)
3. Go to Lighthouse tab
4. Select categories: Performance, Accessibility, Best Practices, SEO
5. Click "Analyze page load"

**Test on these pages:**
- [ ] Homepage
- [ ] Article list page
- [ ] Article detail page

**Target Scores:**
- [ ] Performance: ≥ 90
- [ ] Accessibility: ≥ 95
- [ ] Best Practices: ≥ 90
- [ ] SEO: ≥ 95

---

### 5.2 Core Web Vitals

**Goal:** Verify good user experience metrics

**Check in Lighthouse report:**
- [ ] **LCP** (Largest Contentful Paint): < 2.5s
- [ ] **FID** (First Input Delay): < 100ms (or **INP** < 200ms in newer Chrome)
- [ ] **CLS** (Cumulative Layout Shift): < 0.1

---

### 5.3 DOMPurify Performance Impact

**Goal:** Verify sanitization doesn't slow down page rendering

**Steps:**
1. Open article page
2. Open DevTools → Performance tab
3. Start recording
4. Refresh page
5. Stop recording
6. Analyze timeline

**Success Criteria:**
- [ ] Total page load time < 3 seconds
- [ ] DOMPurify sanitization overhead < 5ms per article
- [ ] No long tasks (> 50ms) caused by sanitization

---

## 6. Functional Testing

### 6.1 Word Count Accuracy

**Goal:** Verify word count calculation is accurate

**Steps:**
1. Login to admin panel
2. Open existing article or create new one
3. Add content with:
   - Nested HTML: `<p>Hello <strong>world</strong></p>`
   - Lists: `<ul><li>First item</li><li>Second item</li></ul>`
   - HTML entities: `It&#39;s working`
   - Line breaks: `<p>Line 1<br>Line 2</p>`
4. Observe word count badge

**Success Criteria:**
- [ ] Word count accurately reflects visible text only
- [ ] HTML tags are not counted
- [ ] Words in nested tags are counted correctly
- [ ] HTML entities are decoded before counting
- [ ] Word count badge has proper color/label:
  - < 300 words: Red "Too short"
  - 300-599: Orange "Short"
  - 600-799: Yellow "Fair"
  - 800-1499: Green "Good"
  - 1500+: Blue "Excellent"
- [ ] Badge has aria-label for screen readers

---

### 6.2 E2E Smoke Tests

**Goal:** Verify basic site functionality works

**Steps:**
```bash
cd web
npm run dev  # Terminal 1
npm run test:e2e -- smoke.spec.ts  # Terminal 2
```

**Success Criteria:**
- [ ] All 6 smoke tests pass
- [ ] Homepage loads successfully
- [ ] Navigation works
- [ ] Can navigate to articles page
- [ ] Search is accessible
- [ ] Politicians page loads
- [ ] Categories page loads

---

## 7. Cross-Browser Testing

**Goal:** Verify site works across major browsers

**Test in these browsers:**
- [ ] Chrome (latest)
- [ ] Firefox (latest)
- [ ] Safari (latest) - macOS/iOS
- [ ] Edge (latest)

**For each browser, verify:**
- [ ] Pages load correctly
- [ ] No JavaScript errors in console
- [ ] Sanitization works (no XSS)
- [ ] Word count displays correctly
- [ ] Schema.org markup renders
- [ ] Accessibility features work (skip link, focus indicators)

---

## 8. Mobile Responsiveness

**Goal:** Verify mobile experience is good

**Steps:**
1. Open site in Chrome DevTools
2. Toggle device toolbar (Ctrl+Shift+M)
3. Test on different viewports:
   - iPhone SE (375px)
   - iPhone 12 Pro (390px)
   - iPad (768px)
   - Desktop (1920px)

**Success Criteria:**
- [ ] Layout adapts to different screen sizes
- [ ] No horizontal scrolling
- [ ] Text is readable (font size ≥ 16px)
- [ ] Tap targets are ≥ 48x48px
- [ ] Images scale properly

---

## 9. Integration Tests

**Goal:** Run full Docker stack integration tests

**Steps:**
```bash
# From project root
docker compose -f docker-compose.prod.yml -f docker-compose.ci.yml up -d
# Wait for services to be healthy
# Run integration tests
docker exec pulpulitiko-api /usr/local/bin/migrate -path /app/migrations -database "postgres://..." up
```

**Or run via GitHub Actions:**
```bash
git push origin feat/Seo
# Check Actions tab on GitHub
```

**Success Criteria:**
- [ ] All services start successfully
- [ ] Database migrations run
- [ ] API health check passes
- [ ] Frontend health check passes
- [ ] E2E tests pass in CI

---

## 10. Deployment Readiness Checklist

Before deploying to production, ensure:

**Code Quality:**
- [ ] All unit tests passing (`npm run test:run` + `go test ./...`)
- [ ] All E2E tests passing (`npm run test:e2e`)
- [ ] No ESLint errors
- [ ] No console.log statements in production code

**Security:**
- [ ] XSS prevention tested and working
- [ ] CSP headers configured
- [ ] No secrets in code or commits
- [ ] Environment variables properly set

**Performance:**
- [ ] Lighthouse scores meet targets
- [ ] Core Web Vitals within limits
- [ ] Images optimized
- [ ] Caching configured

**SEO:**
- [ ] robots.txt working correctly
- [ ] Sitemap generated and accessible
- [ ] Schema.org markup valid
- [ ] Meta tags present on all pages
- [ ] Canonical URLs set correctly

**Accessibility:**
- [ ] WCAG 2.1 AA compliant
- [ ] Keyboard navigation works
- [ ] Screen reader friendly
- [ ] Color contrast sufficient

---

## 11. Post-Deployment Monitoring

**First 24 Hours:**
- [ ] Monitor CSP violation reports (`/api/csp-report`)
- [ ] Check Google Search Console for indexation issues
- [ ] Monitor Core Web Vitals in Search Console
- [ ] Review server logs for sanitization errors
- [ ] Test Rich Results with Google's tool

**First Week:**
- [ ] Monitor search rankings
- [ ] Check for any XSS reports
- [ ] Review performance metrics
- [ ] Verify sitemap indexation progress

---

## Validation Summary

Once all checks are complete, fill out this summary:

**Schema.org & SEO:**
- Google Rich Results: ✅ / ❌
- Schema.org Validator: ✅ / ❌
- robots.txt: ✅ / ❌
- Meta tags: ✅ / ❌

**Accessibility:**
- WAVE scan: ✅ / ❌
- Keyboard navigation: ✅ / ❌
- Automated tests: ✅ / ❌

**Security:**
- XSS prevention: ✅ / ❌
- CSP headers: ✅ / ❌
- Automated tests: ✅ / ❌

**Performance:**
- Lighthouse scores: ✅ / ❌
- Core Web Vitals: ✅ / ❌

**Functional:**
- Word count: ✅ / ❌
- Smoke tests: ✅ / ❌

**Ready for Production:** ✅ / ❌

---

## Support

If any validation fails:
1. Check the relevant test file in `web/tests/e2e/`
2. Review implementation in plan at `~/.claude/plans/federated-plotting-dawn.md`
3. Check browser console for errors
4. Review server logs for backend issues
