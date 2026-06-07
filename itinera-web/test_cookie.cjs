const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();
  
  page.on('response', resp => {
    if (resp.url().includes('/api/v1/trips')) {
      console.log('Response headers:', resp.headers());
    }
  });

  await page.goto('http://localhost:5173/trips');
  await page.waitForLoadState('networkidle');
  
  const cookies = await context.cookies();
  console.log('Cookies:', cookies);
  
  const ls = await page.evaluate(() => window.localStorage.getItem('session_id'));
  console.log('LocalStorage session_id:', ls);
  
  await browser.close();
})();
