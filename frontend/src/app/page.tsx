'use client';

import React, { useState, useEffect, useCallback } from 'react';
import {
  Container,
  Box,
  Typography,
  Button,
  Grid,
  TextField,
  MenuItem,
  InputAdornment,
  CircularProgress,
  Alert,
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import SearchIcon from '@mui/icons-material/Search';
import RefreshIcon from '@mui/icons-material/Refresh';
import { useAuth } from '@/context/AuthContext';
import { Todo, TodoStats, TodoStatus, TodoPriority } from '@/lib/types';
import { api } from '@/lib/api';
import { TodoCard } from '@/components/TodoCard';
import { TodoFormDialog } from '@/components/TodoFormDialog';
import { StatsDashboard } from '@/components/StatsDashboard';
import { useRouter } from 'next/navigation';

export default function HomePage() {
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  const router = useRouter();

  const [todos, setTodos] = useState<Todo[]>([]);
  const [stats, setStats] = useState<TodoStats>({
    TOTAL: 0,
    PENDING: 0,
    IN_PROGRESS: 0,
    COMPLETED: 0,
    CANCELLED: 0,
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filters
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [priorityFilter, setPriorityFilter] = useState<string>('');
  const [search, setSearch] = useState<string>('');

  // Dialog state
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedTodo, setSelectedTodo] = useState<Todo | null>(null);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const [todosRes, statsRes] = await Promise.all([
        api.listTodos({
          status: statusFilter || undefined,
          priority: priorityFilter || undefined,
          search: search.trim() || undefined,
        }),
        api.getTodoStats(),
      ]);
      setTodos(todosRes.data || []);
      setStats(statsRes.stats || { TOTAL: 0, PENDING: 0, IN_PROGRESS: 0, COMPLETED: 0, CANCELLED: 0 });
    } catch (err: any) {
      setError(err.message || 'Failed to fetch tasks');
    } finally {
      setLoading(false);
    }
  }, [statusFilter, priorityFilter, search]);

  useEffect(() => {
    if (!authLoading) {
      if (!isAuthenticated) {
        router.push('/login');
      } else {
        fetchData();
      }
    }
  }, [isAuthenticated, authLoading, fetchData, router]);

  const handleStatusChange = async (id: string, newStatus: TodoStatus, version: number) => {
    try {
      await api.changeStatus(id, newStatus, version);
      fetchData();
    } catch (err: any) {
      alert(`Status transition error: ${err.message}`);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this task?')) return;
    try {
      await api.deleteTodo(id);
      fetchData();
    } catch (err: any) {
      alert(`Deletion error: ${err.message}`);
    }
  };

  const handleOpenCreate = () => {
    setSelectedTodo(null);
    setDialogOpen(true);
  };

  const handleOpenEdit = (todo: Todo) => {
    setSelectedTodo(todo);
    setDialogOpen(true);
  };

  const handleFormSubmit = async (data: {
    title: string;
    description: string;
    priority: TodoPriority;
    due_date?: string;
    tags: string;
    version?: number;
  }) => {
    if (selectedTodo) {
      await api.updateTodo(selectedTodo.id, {
        ...data,
        version: selectedTodo.version,
      });
    } else {
      await api.createTodo(data);
    }
    fetchData();
  };

  if (authLoading || (!isAuthenticated && loading)) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Container maxWidth="lg">
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Box>
          <Typography variant="h4" sx={{ fontWeight: 800 }}>
            Task Workspace
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Domain-driven state machines, optimistic concurrency control & immutable audit logging.
          </Typography>
        </Box>
        <Box sx={{ display: 'flex', gap: 1.5 }}>
          <Button
            variant="outlined"
            startIcon={<RefreshIcon />}
            onClick={() => fetchData()}
            disabled={loading}
          >
            Refresh
          </Button>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={handleOpenCreate}
            sx={{ px: 2.5 }}
          >
            New Task
          </Button>
        </Box>
      </Box>

      {/* KPI Stats */}
      <StatsDashboard stats={stats} />

      {/* Search & Filter Bar */}
      <Box
        sx={{
          p: 2,
          mb: 3,
          bgcolor: 'background.paper',
          borderRadius: 2.5,
          border: '1px solid #e2e8f0',
          display: 'flex',
          flexWrap: 'wrap',
          gap: 2,
          alignItems: 'center',
        }}
      >
        <TextField
          size="small"
          placeholder="Search title, description, or tags..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          sx={{ flexGrow: 1, minWidth: 240 }}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon fontSize="small" color="action" />
              </InputAdornment>
            ),
          }}
        />

        <TextField
          size="small"
          select
          label="Status"
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          sx={{ minWidth: 150 }}
        >
          <MenuItem value="">All Statuses</MenuItem>
          <MenuItem value="PENDING">Pending</MenuItem>
          <MenuItem value="IN_PROGRESS">In Progress</MenuItem>
          <MenuItem value="COMPLETED">Completed</MenuItem>
          <MenuItem value="CANCELLED">Cancelled</MenuItem>
        </TextField>

        <TextField
          size="small"
          select
          label="Priority"
          value={priorityFilter}
          onChange={(e) => setPriorityFilter(e.target.value)}
          sx={{ minWidth: 140 }}
        >
          <MenuItem value="">All Priorities</MenuItem>
          <MenuItem value="LOW">Low</MenuItem>
          <MenuItem value="MEDIUM">Medium</MenuItem>
          <MenuItem value="HIGH">High</MenuItem>
          <MenuItem value="URGENT">Urgent</MenuItem>
        </TextField>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {/* Task List */}
      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
          <CircularProgress />
        </Box>
      ) : todos.length === 0 ? (
        <Box sx={{ textAlign: 'center', py: 10, bgcolor: 'background.paper', borderRadius: 3, border: '1px dashed #cbd5e1' }}>
          <Typography variant="h6" color="text.secondary" sx={{ fontWeight: 600 }}>
            No tasks found
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 1, mb: 2 }}>
            Get started by creating your first task in the domain workspace.
          </Typography>
          <Button variant="contained" startIcon={<AddIcon />} onClick={handleOpenCreate}>
            Create Task
          </Button>
        </Box>
      ) : (
        <Grid container spacing={2}>
          {todos.map((todo) => (
            <Grid item xs={12} sm={6} md={4} key={todo.id}>
              <TodoCard
                todo={todo}
                onStatusChange={handleStatusChange}
                onEdit={handleOpenEdit}
                onDelete={handleDelete}
              />
            </Grid>
          ))}
        </Grid>
      )}

      {/* Create / Edit Dialog */}
      <TodoFormDialog
        open={dialogOpen}
        todo={selectedTodo}
        onClose={() => setDialogOpen(false)}
        onSubmit={handleFormSubmit}
      />
    </Container>
  );
}
