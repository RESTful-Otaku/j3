import type { CapacitorConfig } from '@capacitor/cli';

// J3_SERVER_URL is baked into the native app at sync time. Set it to the LAN
// address of the machine running `bun run demo:phone` (for example
// http://192.168.1.10:8787). When unset the app uses the bundled static UI and
// talks to relative API paths, which is only useful for UI-only QA.
const serverURL = process.env.J3_SERVER_URL || '';

const config: CapacitorConfig = {
  appId: 'com.j3.tournaments',
  appName: 'J3',
  webDir: 'build',
  server: {
    androidScheme: 'http',
    ...(serverURL ? { url: serverURL, cleartext: true, allowNavigation: ['*'] } : {}),
  },
};

export default config;
