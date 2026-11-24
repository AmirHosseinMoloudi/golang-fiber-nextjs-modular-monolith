import React from 'react';
import {
  Card,
  CardContent,
  Typography,
  Box,
  Chip,
  IconButton,
  Button,
  Tooltip,
} from '@mui/material';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import CheckIcon from '@mui/icons-material/Check';
import CancelIcon from '@mui/icons-material/Cancel';
import ReplayIcon from '@mui/icons-material/Replay';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import EditIcon from '@mui/icons-material/Edit';
import EventIcon from '@mui/icons-material/Event';
import LocalOfferIcon from '@mui/icons-material/LocalOffer';
import { Todo, TodoStatus, TodoPriority } from '@/lib/types';

interface TodoCardProps {
  todo: Todo;
  onStatusChange: (id: string, status: TodoStatus, version: number) => void;
  onEdit: (todo: Todo) => void;
  onDelete: (id: string) => void;
}

const statusColors: Record<TodoStatus, 'default' | 'primary' | 'success' | 'error'> = {
  PENDING: 'default',
  IN_PROGRESS: 'primary',
  COMPLETED: 'success',
  CANCELLED: 'error',
};

const priorityColors: Record<TodoPriority, string> = {
  LOW: '#10b981',
  MEDIUM: '#3b82f6',
  HIGH: '#f59e0b',
  URGENT: '#ef4444',
};

export const TodoCard: React.FC<TodoCardProps> = ({
  todo,
  onStatusChange,
  onEdit,
  onDelete,
}) => {
  const isCompleted = todo.status === 'COMPLETED';
  const isCancelled = todo.status === 'CANCELLED';

  return (
    <Card
      sx={{
        transition: 'transform 0.15s ease, box-shadow 0.15s ease',
        '&:hover': {
          transform: 'translateY(-2px)',
          boxShadow: '0 4px 12px rgba(0,0,0,0.08)',
        },
        borderLeft: `5px solid ${priorityColors[todo.priority]}`,
        opacity: isCompleted || isCancelled ? 0.8 : 1,
      }}
    >
      <CardContent sx={{ p: 2.5, '&:last-child': { pb: 2.5 } }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 1 }}>
          <Typography
            variant="h6"
            sx={{
              fontWeight: 700,
              fontSize: '1.05rem',
              textDecoration: isCompleted ? 'line-through' : 'none',
              color: isCompleted ? 'text.secondary' : 'text.primary',
            }}
          >
            {todo.title}
          </Typography>
          <Box sx={{ display: 'flex', gap: 0.5 }}>
            <Tooltip title="Edit Task">
              <IconButton size="small" onClick={() => onEdit(todo)} color="default">
                <EditIcon fontSize="small" />
              </IconButton>
            </Tooltip>
            <Tooltip title="Delete Task">
              <IconButton size="small" onClick={() => onDelete(todo.id)} color="error">
                <DeleteOutlineIcon fontSize="small" />
              </IconButton>
            </Tooltip>
          </Box>
        </Box>

        {todo.description && (
          <Typography variant="body2" color="text.secondary" sx={{ mb: 1.5, whiteSpace: 'pre-line' }}>
            {todo.description}
          </Typography>
        )}

        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, alignItems: 'center', mb: 2 }}>
          <Chip
            label={todo.status.replace('_', ' ')}
            color={statusColors[todo.status]}
            size="small"
            sx={{ fontWeight: 600, fontSize: '0.75rem' }}
          />
          <Chip
            label={todo.priority}
            size="small"
            sx={{
              fontWeight: 700,
              fontSize: '0.7rem',
              bgcolor: `${priorityColors[todo.priority]}15`,
              color: priorityColors[todo.priority],
            }}
          />
          {todo.is_overdue && (
            <Chip
              label="OVERDUE"
              size="small"
              color="error"
              sx={{ fontWeight: 800, fontSize: '0.7rem' }}
            />
          )}
          {todo.due_date && (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, color: 'text.secondary' }}>
              <EventIcon sx={{ fontSize: 16 }} />
              <Typography variant="caption" sx={{ fontWeight: 500 }}>
                {new Date(todo.due_date).toLocaleDateString()}
              </Typography>
            </Box>
          )}
          {todo.tags && (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, color: 'text.secondary' }}>
              <LocalOfferIcon sx={{ fontSize: 14 }} />
              <Typography variant="caption" color="text.secondary">
                {todo.tags}
              </Typography>
            </Box>
          )}
          <Chip
            label={`v${todo.version}`}
            size="small"
            variant="outlined"
            sx={{ height: 20, fontSize: '0.65rem', color: 'text.secondary', ml: 'auto' }}
          />
        </Box>

        {/* State Machine Transition Actions */}
        <Box sx={{ display: 'flex', gap: 1, pt: 1, borderTop: '1px solid #f1f5f9' }}>
          {todo.status === 'PENDING' && (
            <>
              <Button
                size="small"
                variant="outlined"
                startIcon={<PlayArrowIcon />}
                onClick={() => onStatusChange(todo.id, 'IN_PROGRESS', todo.version)}
              >
                Start
              </Button>
              <Button
                size="small"
                variant="outlined"
                color="success"
                startIcon={<CheckIcon />}
                onClick={() => onStatusChange(todo.id, 'COMPLETED', todo.version)}
              >
                Complete
              </Button>
              <Button
                size="small"
                color="error"
                startIcon={<CancelIcon />}
                onClick={() => onStatusChange(todo.id, 'CANCELLED', todo.version)}
              >
                Cancel
              </Button>
            </>
          )}

          {todo.status === 'IN_PROGRESS' && (
            <>
              <Button
                size="small"
                variant="contained"
                color="success"
                startIcon={<CheckIcon />}
                onClick={() => onStatusChange(todo.id, 'COMPLETED', todo.version)}
              >
                Complete
              </Button>
              <Button
                size="small"
                color="error"
                startIcon={<CancelIcon />}
                onClick={() => onStatusChange(todo.id, 'CANCELLED', todo.version)}
              >
                Cancel
              </Button>
            </>
          )}

          {(isCompleted || isCancelled) && (
            <Button
              size="small"
              variant="outlined"
              color="primary"
              startIcon={<ReplayIcon />}
              onClick={() => onStatusChange(todo.id, 'PENDING', todo.version)}
            >
              Reopen
            </Button>
          )}
        </Box>
      </CardContent>
    </Card>
  );
};
