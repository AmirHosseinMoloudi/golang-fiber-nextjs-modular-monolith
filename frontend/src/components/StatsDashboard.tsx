import React from 'react';
import { Grid, Paper, Typography, Box } from '@mui/material';
import AssignmentIcon from '@mui/icons-material/Assignment';
import HourglassEmptyIcon from '@mui/icons-material/HourglassEmpty';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import HighlightOffIcon from '@mui/icons-material/HighlightOff';
import { TodoStats } from '@/lib/types';

interface StatsDashboardProps {
  stats: TodoStats;
}

export const StatsDashboard: React.FC<StatsDashboardProps> = ({ stats }) => {
  const items = [
    { label: 'Total Tasks', value: stats.TOTAL || 0, color: '#2563eb', icon: <AssignmentIcon sx={{ color: '#2563eb' }} /> },
    { label: 'Pending', value: stats.PENDING || 0, color: '#64748b', icon: <HourglassEmptyIcon sx={{ color: '#64748b' }} /> },
    { label: 'In Progress', value: stats.IN_PROGRESS || 0, color: '#f59e0b', icon: <HourglassEmptyIcon sx={{ color: '#f59e0b' }} /> },
    { label: 'Completed', value: stats.COMPLETED || 0, color: '#16a34a', icon: <CheckCircleIcon sx={{ color: '#16a34a' }} /> },
    { label: 'Cancelled', value: stats.CANCELLED || 0, color: '#ef4444', icon: <HighlightOffIcon sx={{ color: '#ef4444' }} /> },
  ];

  return (
    <Grid container spacing={2} sx={{ mb: 4 }}>
      {items.map((item, idx) => (
        <Grid item xs={6} sm={4} md={2.4} key={idx}>
          <Paper
            sx={{
              p: 2,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              borderRadius: 2.5,
              transition: 'box-shadow 0.2s',
              '&:hover': {
                boxShadow: '0 4px 12px rgba(0,0,0,0.06)',
              },
            }}
          >
            <Box>
              <Typography variant="caption" color="text.secondary" sx={{ fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                {item.label}
              </Typography>
              <Typography variant="h4" sx={{ fontWeight: 800, mt: 0.5, color: item.color }}>
                {item.value}
              </Typography>
            </Box>
            <Box sx={{ p: 1, borderRadius: 2, bgcolor: `${item.color}15` }}>
              {item.icon}
            </Box>
          </Paper>
        </Grid>
      ))}
    </Grid>
  );
};
