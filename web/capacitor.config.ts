import type { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
  appId: 'com.j3.tournaments',
  appName: 'J3',
  webDir: 'build',
  server: {
    // Local venue server runs on the LAN; override at build time for a real deployment.
    androidScheme: 'http',
  },
};

export default config;
