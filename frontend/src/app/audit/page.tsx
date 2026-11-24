'use client';

import React, { useState, useEffect, useCallback } from 'react';
import {
  Container,
  Paper,
  Box,
  Typography,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  CircularProgress,
  Alert,
} from '@mui/material';
import VerifiedUserIcon from '@mui/icons-material/VerifiedUser';
import RefreshIcon from '@mui/icons-material/Refresh';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import { useAuth } from '@/context/AuthContext';
import { AuditRecord } from '@/lib/types';
import { api } from '@/lib/api';
import { useRouter } from 'next/navigation';
import Link from 'next/link';

export default function AuditPage() {
  const { isAuditor, isAuthenticated, isLoading: authLoading } = useAuth();
  const router = useRouter();

  const [records, setRecords] = useState<AuditRecord[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [verifying, setVerifying] = useState(false);
  const [integrityStatus, setIntegrityStatus] = useState<boolean | null>(null);
  const [error, setError] = useState<string | null>(null);

  const fetchAuditLogs = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await api.listAuditLogs(1);
      setRecords(res.data || []);
      setTotal(res.total || 0);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch audit records');
    } finally {
      setLoading(false);
    }
  }, []);

  const handleVerifyIntegrity = async () => {
    try {
      setVerifying(true);
      const res = await api.verifyAuditIntegrity();
      setIntegrityStatus(res.tamper_evident_integrity_valid);
    } catch (err: any) {
      alert(`Verification failed: ${err.message}`);
    } finally {
      setVerifying(false);
    }
  };

  useEffect(() => {
    if (!authLoading) {
      if (!isAuthenticated) {
        router.push('/login');
      } else if (!isAuditor) {
        router.push('/');
      } else {
        fetchAuditLogs();
      }
    }
  }, [authLoading, isAuthenticated, isAuditor, fetchAuditLogs, router]);

  if (authLoading || loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Container maxWidth="xl">
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <Button component={Link} href="/" startIcon={<ArrowBackIcon />} color="inherit">
            Back
          </Button>
          <Box>
            <Typography variant="h4" sx={{ fontWeight: 800 }}>
              Tamper-Evident Audit Trail
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Cryptographically chained append-only ledger with SHA-256 integrity proofs.
            </Typography>
          </Box>
        </Box>

        <Box sx={{ display: 'flex', gap: 1.5 }}>
          <Button
            variant="outlined"
            startIcon={<RefreshIcon />}
            onClick={() => fetchAuditLogs()}
          >
            Refresh
          </Button>
          <Button
            variant="contained"
            color="success"
            startIcon={<VerifiedUserIcon />}
            onClick={handleVerifyIntegrity}
            disabled={verifying}
          >
            {verifying ? 'Verifying...' : 'Verify Chain Integrity'}
          </Button>
        </Box>
      </Box>

      {integrityStatus !== null && (
        <Alert
          severity={integrityStatus ? 'success' : 'error'}
          sx={{ mb: 3 }}
        >
          {integrityStatus
            ? 'Cryptographic Hash Chain Verified: 100% of audit ledger blocks are tamper-free and match the genesis state.'
            : 'Warning: Hash chain integrity verification failed! Discrepancy detected in ledger blocks.'}
        </Alert>
      )}

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      <TableContainer component={Paper} sx={{ borderRadius: 2.5 }}>
        <Table sx={{ minWidth: 650 }}>
          <TableHead sx={{ bgcolor: '#f8fafc' }}>
            <TableRow>
              <TableCell sx={{ fontWeight: 700 }}>Timestamp</TableCell>
              <TableCell sx={{ fontWeight: 700 }}>Action</TableCell>
              <TableCell sx={{ fontWeight: 700 }}>Entity</TableCell>
              <TableCell sx={{ fontWeight: 700 }}>Actor</TableCell>
              <TableCell sx={{ fontWeight: 700 }}>Payload Snapshot</TableCell>
              <TableCell sx={{ fontWeight: 700 }}>SHA-256 Hash</TableCell>
              <TableCell sx={{ fontWeight: 700 }}>Previous Hash</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {records.map((r) => (
              <TableRow key={r.id} hover>
                <TableCell sx={{ whiteSpace: 'nowrap' }}>
                  {new Date(r.timestamp).toLocaleString()}
                </TableCell>
                <TableCell>
                  <Chip
                    label={r.action}
                    size="small"
                    color={
                      r.action === 'CREATED'
                        ? 'success'
                        : r.action === 'DELETED'
                        ? 'error'
                        : r.action === 'STATUS_CHANGED'
                        ? 'primary'
                        : 'default'
                    }
                    sx={{ fontWeight: 700, fontSize: '0.7rem' }}
                  />
                </TableCell>
                <TableCell>
                  <Typography variant="body2" sx={{ fontWeight: 600 }}>
                    {r.entity_type}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    {r.entity_id.substring(0, 8)}...
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="body2">{r.actor_email || 'System'}</Typography>
                </TableCell>
                <TableCell sx={{ maxWidth: 320 }}>
                  <Typography
                    variant="caption"
                    sx={{
                      fontFamily: 'monospace',
                      display: '-webkit-box',
                      WebkitLineClamp: 2,
                      WebkitBoxOrient: 'vertical',
                      overflow: 'hidden',
                      bgcolor: '#f1f5f9',
                      p: 0.5,
                      borderRadius: 1,
                    }}
                  >
                    {r.payload}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="caption" sx={{ fontFamily: 'monospace', color: 'primary.main', fontWeight: 600 }}>
                    {r.record_hash.substring(0, 12)}...
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="caption" sx={{ fontFamily: 'monospace', color: 'text.secondary' }}>
                    {r.previous_hash.substring(0, 12)}...
                  </Typography>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
      <Box sx={{ mt: 2, textAlign: 'right' }}>
        <Typography variant="caption" color="text.secondary">
          Showing {records.length} of {total} records
        </Typography>
      </Box>
    </Container>
  );
}
