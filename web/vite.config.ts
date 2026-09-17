import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const backendPort = env.VITE_BACKEND_PORT || env.BACKEND_PORT || '8080';
  const backendTarget = env.VITE_BACKEND_URL || `http://localhost:${backendPort}`;

  return {
    plugins: [react()],
    server: {
      port: 3000,
      proxy: {
        '/api': {
          target: backendTarget,
          changeOrigin: true,
          secure: false,
          router: (req) => {
            const customTarget = req.headers['x-backend-url'] as string;
            if (customTarget && customTarget.startsWith('http')) {
              return customTarget;
            }
            return backendTarget;
          },
        },
        '/graphql': {
          target: backendTarget,
          changeOrigin: true,
          secure: false,
          router: (req) => {
            const customTarget = req.headers['x-backend-url'] as string;
            if (customTarget && customTarget.startsWith('http')) {
              return customTarget;
            }
            return backendTarget;
          },
        },
      },
    },
    build: {
      outDir: 'dist',
    },
  };
});
