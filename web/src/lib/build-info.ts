// Small, visible build marker so device QA can tell local, QA and production
// bundles apart without opening devtools. Flavours are set by build scripts.
export const buildFlavor =
  import.meta.env.VITE_BUILD_FLAVOR ||
  (import.meta.env.VITE_OFFLINE_DEMO === '1' ? 'qa-offline' : 'local');
