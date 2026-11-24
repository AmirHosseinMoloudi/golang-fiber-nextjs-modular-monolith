'use client';

import React from 'react';
import {
  AppBar,
  Toolbar,
  Typography,
  Button,
  Box,
  Chip,
  IconButton,
  Tooltip,
} from '@mui/material';
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline';
import SecurityIcon from '@mui/icons-material/Security';
import LogoutIcon from '@mui/icons-material/Logout';
import AssessmentIcon from '@mui/icons-material/Assessment';
import { useAuth } from '@/context/AuthContext';
import Link from 'next/link';

export const Navbar: React.FC = () => {
  const { user, isAuthenticated, isAuditor, logout } = useAuth();

  return (
    <AppBar position="sticky" color="inherit" elevation={0} sx={{ borderBottom: '1px solid #e2e8f0' }}>
      <Toolbar sx={{ justifyContent: 'space-between', px: { xs: 2, md: 4 } }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
          <CheckCircleOutlineIcon color="primary" sx={{ fontSize: 32 }} />
          <Typography
            variant="h6"
            component={Link}
            href="/"
            sx={{
              fontWeight: 800,
              color: 'text.primary',
              textDecoration: 'none',
              background: 'linear-gradient(45deg, #2563eb, #7c3aed)',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
            }}
          >
            OmniTask
          </Typography>
          <Chip label="v1.0 DDD" size="small" variant="outlined" color="primary" sx={{ ml: 1, height: 22, fontSize: '0.75rem' }} />
        </Box>

        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          {isAuthenticated ? (
            <>
              <Button component={Link} href="/" color="inherit" sx={{ fontWeight: 600 }}>
                Tasks
              </Button>
              {isAuditor && (
                <Button
                  component={Link}
                  href="/audit"
                  startIcon={<SecurityIcon />}
                  color="inherit"
                  sx={{ fontWeight: 600 }}
                >
                  Audit Trail
                </Button>
              )}
              <Button
                component="a"
                href="http://localhost:3001"
                target="_blank"
                rel="noreferrer"
                startIcon={<AssessmentIcon />}
                color="inherit"
                sx={{ fontWeight: 600 }}
              >
                Grafana
              </Button>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, ml: 1 }}>
                <Chip
                  label={user?.role}
                  color={user?.role === 'ADMIN' ? 'error' : user?.role === 'AUDITOR' ? 'warning' : 'info'}
                  size="small"
                  sx={{ fontWeight: 700, fontSize: '0.7rem' }}
                />
                <Typography variant="body2" color="text.secondary" sx={{ display: { xs: 'none', sm: 'block' }, fontWeight: 500 }}>
                  {user?.full_name}
                </Typography>
                <Tooltip title="Logout">
                  <IconButton onClick={logout} size="small" color="default" sx={{ ml: 0.5 }}>
                    <LogoutIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              </Box>
            </>
          ) : (
            <>
              <Button component={Link} href="/login" color="inherit">
                Login
              </Button>
              <Button component={Link} href="/register" variant="contained" color="primary">
                Register
              </Button>
            </>
          )}
        </Box>
      </Toolbar>
    </AppBar>
  );
};
