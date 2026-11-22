import type { CapacitorConfig } from '@capacitor/cli';

// The native WebView always loads the bundled Svelte app. The venue server URL
// is baked into the JS bundle as VITE_API_BASE_URL at build time (see
// scripts/mobile.sh), so API calls reach the LAN server without making the
// WebView navigate away from local assets.
const config: CapacitorConfig = {
  appId: 'com.j3.tournaments',
  appName: 'J3',
  webDir: 'build',
  server: {
    androidScheme: 'http',
  },
};

export default config;
