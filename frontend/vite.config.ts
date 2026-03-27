import preact from '@preact/preset-vite';
import { readFileSync, writeFileSync, existsSync } from 'fs';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';
import { defineConfig, type Plugin } from 'vite';

const __dirname = dirname(fileURLToPath(import.meta.url));

// GitHub Pages project sites live under /<repo>/; CI sets VITE_BASE_PATH (see deploy workflow).
// Local dev and builds without VITE_BASE_PATH keep base "/".
const base =
  process.env.VITE_BASE_PATH && process.env.VITE_BASE_PATH !== '/'
    ? process.env.VITE_BASE_PATH.replace(/\/?$/, '/')
    : '/';

/** Strip trailing slash; empty if unset. Used for canonical, sitemap, and JSON-LD. */
function normalizedSiteUrl(): string {
  const raw = process.env.VITE_SITE_URL?.trim();
  if (!raw) return '';
  return raw.replace(/\/+$/, '');
}

function seoBuildPlugin(siteRoot: string): Plugin {
  return {
    name: 'seo-build',
    transformIndexHtml(html) {
      if (!siteRoot) {
        return html.replace(/<!-- SEO:ABSOLUTE_START -->[\s\S]*?<!-- SEO:ABSOLUTE_END -->/g, '');
      }
      return html
        .replace(/__SITE_URL__/g, siteRoot)
        .replace(/<!-- SEO:ABSOLUTE_START -->\n?/g, '')
        .replace(/\n?<!-- SEO:ABSOLUTE_END -->\n?/g, '\n');
    },
    closeBundle() {
      if (!siteRoot) return;
      const dist = resolve(__dirname, 'dist');
      const sitemap = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>${siteRoot}/</loc>
    <changefreq>weekly</changefreq>
    <priority>1.0</priority>
  </url>
</urlset>
`;
      writeFileSync(resolve(dist, 'sitemap.xml'), sitemap);
      const robotsPath = resolve(dist, 'robots.txt');
      if (existsSync(robotsPath)) {
        const existing = readFileSync(robotsPath, 'utf8');
        if (!/^Sitemap:\s/im.test(existing)) {
          writeFileSync(
            robotsPath,
            `${existing.replace(/\s+$/, '')}\n\nSitemap: ${siteRoot}/sitemap.xml\n`,
          );
        }
      }
    },
  };
}

const siteUrl = normalizedSiteUrl();

export default defineConfig({
  base,
  plugins: [preact(), seoBuildPlugin(siteUrl)],
});
