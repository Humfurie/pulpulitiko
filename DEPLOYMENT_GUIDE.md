# Deployment Guide

This guide covers deploying the security, SEO, and accessibility improvements to production.

## Pre-Deployment Checklist

Before deploying, ensure all validation passes:

```bash
# Run all tests
cd api && go test ./...
cd ../web && npm run test:run
npm run test:e2e  # (requires dev server)

# Check validation checklist
cat VALIDATION_CHECKLIST.md
```

**Critical Pre-Flight Checks:**
- [ ] All unit tests passing (119 frontend + 37 backend)
- [ ] All E2E tests passing (50 tests)
- [ ] Manual validation checklist completed
- [ ] No secrets in code or environment files
- [ ] Environment variables configured for production
- [ ] Database backup completed
- [ ] Rollback plan prepared

---

## Deployment Strategy

**Recommended Approach:** Blue-Green Deployment

1. Deploy backend first (API with HTML sanitization)
2. Test backend independently
3. Deploy frontend (with DOMPurify + text utilities)
4. Test full stack
5. Monitor for 24 hours before marking complete

**Why Backend First?**
- Backend sanitizes new content at storage time
- Frontend provides additional display-time protection
- If frontend deployment fails, backend protection is still active

---

## Step-by-Step Deployment

### 1. Backup Current State

```bash
# Backup database
docker exec pulpulitiko-postgres pg_dump -U politics politics_db > backup_$(date +%Y%m%d_%H%M%S).sql

# Tag current git commit (before deploying new changes)
git tag pre-seo-security-deploy
git push origin pre-seo-security-deploy

# Backup current Docker images (optional)
docker save pulpulitiko-api:latest > api_backup.tar
docker save pulpulitiko-web:latest > web_backup.tar
```

---

### 2. Deploy Backend (Go API)

**2.1 Review Backend Changes**

Key files changed:
- `api/internal/services/html_sanitizer.go` (NEW)
- `api/internal/services/html_sanitizer_test.go` (NEW)
- `api/internal/services/article_service.go` (MODIFIED)
- `api/internal/services/comment_service.go` (MODIFIED)

**2.2 Build and Deploy Backend**

```bash
# Navigate to project root
cd /home/humfurie/Desktop/Projects/pulpulitiko

# Rebuild API Docker image
docker compose -f docker-compose.prod.yml build api

# Stop current API container
docker stop pulpulitiko-api

# Start new API container
docker compose -f docker-compose.prod.yml up -d api

# Wait for health check
sleep 10
curl -f http://localhost:8080/health || echo "API health check failed!"
```

**2.3 Verify Backend Deployment**

```bash
# Check API logs
docker logs pulpulitiko-api --tail 50

# Test API endpoints
curl http://localhost:8080/api/articles
curl http://localhost:8080/api/politicians

# Verify database connection
docker exec pulpulitiko-api /app/seed --help || echo "Binary works"
```

**Success Criteria:**
- [ ] API container running
- [ ] Health endpoint returns 200
- [ ] No errors in logs
- [ ] Public endpoints return data
- [ ] Database queries work

**If Backend Deployment Fails:**
```bash
# Rollback to previous image
docker compose -f docker-compose.prod.yml down api
docker load < api_backup.tar
docker compose -f docker-compose.prod.yml up -d api
```

---

### 3. Deploy Frontend (Nuxt/Vue)

**3.1 Review Frontend Changes**

Key files changed:
- `web/app/composables/useSanitizedHtml.ts` (NEW)
- `web/app/composables/useTextUtils.ts` (NEW)
- `web/app/pages/article/[slug].vue` (MODIFIED)
- `web/app/components/CommentItem.vue` (MODIFIED)
- `web/app/components/CommentForm.vue` (MODIFIED)
- `web/app/pages/voter-education/[slug].vue` (MODIFIED)
- `web/app/pages/politician/[slug].vue` (MODIFIED)
- `web/app/pages/admin/articles/[id].vue` (MODIFIED)
- `web/app/pages/admin/articles/new.vue` (MODIFIED)
- `web/server/routes/robots.txt.ts` (NEW)
- `web/server/plugins/csp.ts` (NEW)
- `web/server/api/csp-report.post.ts` (NEW)
- `web/vitest.config.ts` (MODIFIED)
- `web/public/robots.txt` (DELETED)

**3.2 Build and Deploy Frontend**

```bash
# Rebuild web Docker image
docker compose -f docker-compose.prod.yml build web

# Stop current web container
docker stop pulpulitiko-web

# Start new web container
docker compose -f docker-compose.prod.yml up -d web

# Wait for startup
sleep 15
curl -f http://localhost:3000 || echo "Frontend health check failed!"
```

**3.3 Verify Frontend Deployment**

```bash
# Check web logs
docker logs pulpulitiko-web --tail 50

# Test homepage
curl -I http://localhost:3000

# Test robots.txt
curl http://localhost:3000/robots.txt

# Test article page
curl -I http://localhost:3000/articles
```

**Success Criteria:**
- [ ] Web container running
- [ ] Homepage loads (HTTP 200)
- [ ] No JavaScript errors in browser console
- [ ] robots.txt returns correct content
- [ ] CSP headers present
- [ ] Article pages load correctly

**If Frontend Deployment Fails:**
```bash
# Rollback to previous image
docker compose -f docker-compose.prod.yml down web
docker load < web_backup.tar
docker compose -f docker-compose.prod.yml up -d web
```

---

### 4. Post-Deployment Verification

**4.1 Critical Functionality Tests**

Access via browser: `https://pulpulitiko.humfurie.org`

**Test Flow:**
1. Navigate to homepage
2. Click on an article
3. Verify article content displays correctly
4. Check if comments display (if enabled)
5. Test search functionality
6. Navigate to politician page
7. Navigate to voter education page

**Check Browser Console:**
- [ ] No JavaScript errors
- [ ] No CSP violations (or only expected ones)
- [ ] No 404 errors for resources

**4.2 Schema.org Validation**

```bash
# Test robots.txt
curl https://pulpulitiko.humfurie.org/robots.txt

# Expected output:
# - User-agent: *
# - Sitemap: https://pulpulitiko.humfurie.org/sitemap.xml
# - NOT localhost
```

**4.3 Security Verification**

**XSS Prevention:**
1. Login to admin (if you have access)
2. Create test article with this content:
   ```html
   <p>Safe content</p>
   <script>alert('XSS')</script>
   <img src=x onerror=alert('XSS')>
   ```
3. Save and view article
4. Verify: No alert boxes, script tags removed

**CSP Headers:**
1. Open browser DevTools → Network
2. Refresh page
3. Check Response Headers for:
   - `Content-Security-Policy` or `Content-Security-Policy-Report-Only`
   - `X-Content-Type-Options: nosniff`
   - `X-Frame-Options`

**4.4 Performance Check**

```bash
# Test response times
time curl -s https://pulpulitiko.humfurie.org > /dev/null
time curl -s https://pulpulitiko.humfurie.org/articles > /dev/null
```

**Expected:** < 3 seconds per request

---

## 5. Monitoring and Observability

### 5.1 Application Logs

**Monitor for these patterns:**

```bash
# Backend logs (watch for sanitization activity)
docker logs -f pulpulitiko-api | grep -i "sanitiz"

# Frontend logs (watch for CSP violations)
docker logs -f pulpulitiko-web | grep -i "csp"

# Error logs
docker logs pulpulitiko-api | grep -i "error"
docker logs pulpulitiko-web | grep -i "error"
```

**Set up log aggregation (optional):**
- Use Loki/Grafana for centralized logging
- Forward logs to CloudWatch, Datadog, or similar

### 5.2 CSP Violation Monitoring

CSP violations are logged to `/api/csp-report` endpoint.

**Monitor violations:**
```bash
# Check web server logs for CSP reports
docker logs pulpulitiko-web | grep "CSP Violation"

# If violations are frequent, review CSP policy
# File: web/server/plugins/csp.ts
```

**Common legitimate violations:**
- Browser extensions (AdBlock, etc.)
- Third-party analytics (if added later)
- Development tools

**Adjust CSP policy if needed:**
1. Identify violation source from logs
2. Evaluate if source is legitimate
3. Update `web/server/plugins/csp.ts` if necessary
4. Redeploy frontend

### 5.3 Error Monitoring

**Set up error tracking (recommended):**
- [Sentry](https://sentry.io) - Application error tracking
- [LogRocket](https://logrocket.com) - Session replay
- [Rollbar](https://rollbar.com) - Error monitoring

**Key metrics to track:**
- Backend: Sanitization errors, database errors
- Frontend: JavaScript errors, CSP violations
- Performance: Response times, Core Web Vitals

---

## 6. Search Engine Monitoring

### 6.1 Google Search Console

**Day 1-3 After Deployment:**
1. Go to [Google Search Console](https://search.google.com/search-console)
2. Check "Coverage" section for errors
3. Submit sitemap (if not already done)
4. Monitor indexation progress

**Week 1 Checklist:**
- [ ] No new indexation errors
- [ ] Sitemap processed successfully
- [ ] Rich results eligible count increases
- [ ] No Core Web Vitals issues

### 6.2 Rich Results Monitoring

**Weekly Check:**
1. Go to [Google Rich Results Test](https://search.google.com/test/rich-results)
2. Test 5 random article URLs
3. Verify NewsArticle and BreadcrumbList schemas detected
4. Check for any new errors or warnings

---

## 7. Performance Monitoring

### 7.1 Core Web Vitals

**Monitor in Google Search Console:**
1. Go to "Core Web Vitals" section
2. Check mobile and desktop metrics
3. Identify any "poor" URLs
4. Investigate and fix performance issues

**Target Metrics:**
- LCP (Largest Contentful Paint): < 2.5s
- FID/INP (First Input Delay / Interaction to Next Paint): < 100ms / 200ms
- CLS (Cumulative Layout Shift): < 0.1

### 7.2 Lighthouse CI (Optional)

**Set up automated Lighthouse audits:**

```yaml
# Add to .github/workflows/lighthouse.yml
name: Lighthouse CI
on: [push]
jobs:
  lighthouse:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: treosh/lighthouse-ci-action@v10
        with:
          urls: |
            https://pulpulitiko.humfurie.org
            https://pulpulitiko.humfurie.org/articles
          uploadArtifacts: true
```

---

## 8. Rollback Procedures

### 8.1 Quick Rollback (< 5 minutes)

If critical issues are discovered immediately after deployment:

```bash
# Stop current containers
docker compose -f docker-compose.prod.yml down

# Restore from backup images
docker load < api_backup.tar
docker load < web_backup.tar

# Start previous version
docker compose -f docker-compose.prod.yml up -d

# Verify services
curl http://localhost:8080/health
curl http://localhost:3000
```

### 8.2 Git Rollback

If you need to revert code changes:

```bash
# Find the commit before deployment
git log --oneline

# Revert to previous tag
git checkout pre-seo-security-deploy

# Or revert specific commits
git revert <commit-hash>

# Rebuild and redeploy
docker compose -f docker-compose.prod.yml build
docker compose -f docker-compose.prod.yml up -d
```

### 8.3 Database Rollback (if needed)

If database migrations were run and need to be reversed:

```bash
# Restore database from backup
docker exec -i pulpulitiko-postgres psql -U politics politics_db < backup_20250110_000000.sql

# Or run down migrations
docker exec pulpulitiko-api /usr/local/bin/migrate \
  -path /app/migrations \
  -database "postgres://politics:password@postgres:5432/politics_db?sslmode=disable" \
  down 1
```

**⚠️ WARNING:** Database rollbacks can cause data loss. Only do this if absolutely necessary and you have a verified backup.

---

## 9. Deployment Timeline

**Recommended Schedule:**

**Day 0 (Pre-Deployment):**
- Complete all validation checks
- Create backups
- Schedule maintenance window (optional)
- Notify team/users if downtime expected

**Day 1 (Deployment):**
- 09:00 - Deploy backend
- 09:30 - Verify backend health
- 10:00 - Deploy frontend
- 10:30 - Verify frontend health
- 11:00 - Run smoke tests
- 12:00 - Monitor for issues
- 17:00 - End of business check

**Day 2-7 (Monitoring):**
- Daily: Check error logs
- Daily: Monitor CSP violations
- Daily: Review performance metrics
- Daily: Check Google Search Console

**Week 2:**
- Run full validation checklist again
- Review any user feedback
- Optimize CSP policy if needed
- Plan next improvements

---

## 10. Known Issues and Workarounds

### Issue: CSP Blocks Inline Styles

**Symptom:** Some styles don't apply due to CSP blocking inline styles

**Workaround:** CSP is currently in report-only mode. Monitor violations for 1 week before enabling enforcement.

**File:** `web/server/plugins/csp.ts`

---

### Issue: MinIO Images Not Loading

**Symptom:** Article images return 403 or CORS errors

**Verification:**
```bash
# Check MinIO CSP whitelist
grep minio.humfurie.org web/server/plugins/csp.ts
```

**Fix:** Ensure MinIO domain is in CSP `img-src` directive

---

### Issue: Word Count Calculation Slow

**Symptom:** Admin article editor lags when typing

**Verification:** Check if word count updates debounce properly

**Optimization:** Word count calculation should be debounced (wait 500ms after typing stops)

---

## 11. Communication Plan

### User Communication

**Before Deployment:**
```
Subject: Upcoming Platform Improvements

We'll be deploying security and SEO improvements on [DATE].

What's changing:
- Enhanced content security (XSS prevention)
- Improved search engine optimization
- Better accessibility features

Expected downtime: None (rolling deployment)
If you notice any issues, please contact [SUPPORT EMAIL]
```

**After Deployment:**
```
Subject: Platform Improvements Deployed Successfully

We've successfully deployed our latest improvements:

✅ Enhanced security measures
✅ Improved SEO for better search rankings
✅ Better accessibility for all users

If you encounter any issues, please report them to [SUPPORT EMAIL]
```

---

## 12. Success Metrics

Track these metrics to measure deployment success:

**Week 1:**
- [ ] Zero critical errors in logs
- [ ] < 1% CSP violation rate
- [ ] Core Web Vitals in "good" range
- [ ] All manual validation checks pass

**Month 1:**
- [ ] Organic search traffic increase (baseline vs. post-deploy)
- [ ] Google Search Console impressions increase
- [ ] Rich results clicks increase
- [ ] Bounce rate stable or decreased
- [ ] Page load time stable or improved

**Quarter 1:**
- [ ] Search ranking improvements for target keywords
- [ ] Accessibility complaints decreased
- [ ] Security incidents: zero XSS attacks
- [ ] SEO score maintained > 95

---

## 13. Deployment Checklist Summary

**Pre-Deployment:**
- [ ] All tests passing
- [ ] Manual validation complete
- [ ] Backups created
- [ ] Environment variables verified
- [ ] Team notified

**Backend Deployment:**
- [ ] Docker image built
- [ ] Container started
- [ ] Health check passing
- [ ] Logs reviewed
- [ ] API endpoints tested

**Frontend Deployment:**
- [ ] Docker image built
- [ ] Container started
- [ ] Homepage loading
- [ ] robots.txt correct
- [ ] CSP headers present
- [ ] No console errors

**Post-Deployment:**
- [ ] Full site tested
- [ ] Schema.org validated
- [ ] XSS prevention verified
- [ ] Performance checked
- [ ] Monitoring configured
- [ ] Team/users notified

**Day 1 Monitoring:**
- [ ] Error logs reviewed (hourly)
- [ ] CSP violations reviewed
- [ ] Performance metrics checked
- [ ] User reports reviewed

**Week 1 Follow-up:**
- [ ] Google Search Console checked
- [ ] Analytics reviewed
- [ ] Validation checklist re-run
- [ ] Issues documented and fixed

---

## 14. Support and Troubleshooting

**If something goes wrong:**

1. **Check logs first:**
   ```bash
   docker logs pulpulitiko-api --tail 100
   docker logs pulpulitiko-web --tail 100
   ```

2. **Check service health:**
   ```bash
   docker ps
   curl http://localhost:8080/health
   curl http://localhost:3000
   ```

3. **Review recent changes:**
   ```bash
   git log --oneline -10
   git diff HEAD~1
   ```

4. **Test in isolation:**
   - Stop frontend, test backend only
   - Check database connectivity
   - Verify environment variables

5. **If all else fails:**
   - Execute rollback procedure (Section 8)
   - Restore from backups
   - Report issue with logs

**Contact Information:**
- Technical Issues: [Your contact]
- Security Concerns: [Security contact]
- Performance Issues: [Performance contact]

---

## Conclusion

This deployment brings critical security improvements (XSS prevention), SEO enhancements (Schema.org, robots.txt), and accessibility features to Pulpulitiko.

**Key Achievements:**
- ✅ Dual-layer HTML sanitization (backend + frontend)
- ✅ Enhanced Schema.org structured data
- ✅ Dynamic robots.txt for all environments
- ✅ Accurate word count calculation
- ✅ CSP headers for additional security
- ✅ WCAG 2.1 AA accessibility compliance
- ✅ Comprehensive test coverage (156 tests)

**Next Steps:**
1. Complete deployment following this guide
2. Monitor for 1 week
3. Run validation checklist
4. Review analytics and Search Console data
5. Plan future improvements

Good luck with your deployment! 🚀
