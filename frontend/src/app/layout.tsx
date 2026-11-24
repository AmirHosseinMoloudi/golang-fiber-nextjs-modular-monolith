import type { Metadata } from 'next';
import ThemeRegistry from '@/theme/ThemeRegistry';
import { AuthProvider } from '@/context/AuthContext';
import { Navbar } from '@/components/Navbar';
import { Box } from '@mui/material';

export const metadata: Metadata = {
  title: 'OmniTask - Modular Monolith DDD Todo Platform',
  description: 'Enterprise production-grade Todo platform built with Go Fiber, GORM, SQLite, Next.js, MUI, and full OpenTelemetry stack.',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>
        <ThemeRegistry>
          <AuthProvider>
            <Box sx={{ minHeight: '100vh', display: 'flex', flexDirection: 'column', bgcolor: 'background.default' }}>
              <Navbar />
              <Box component="main" sx={{ flexGrow: 1, py: 4 }}>
                {children}
              </Box>
            </Box>
          </AuthProvider>
        </ThemeRegistry>
      </body>
    </html>
  );
}
