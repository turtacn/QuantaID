import React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { DataGrid, GridColDef } from '@mui/x-data-grid';
import { Button, Dialog, DialogTitle, DialogContent, TextField } from '@mui/material';
import { userAPI } from '../services/api';

export const UserManagement: React.FC = () => {
  const [page, setPage] = React.useState(0);
  const [search, setSearch] = React.useState('');
  const [openDialog, setOpenDialog] = React.useState(false);
  const queryClient = useQueryClient();

  // 查询用户列表
  const { data: users, isLoading } = useQuery({
    queryKey: ['users', page, search],
    queryFn: () => userAPI.list({ page, size: 20, search }),
  });

  // 创建用户 Mutation
  const createUserMutation = useMutation({
    mutationFn: userAPI.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setOpenDialog(false);
    },
  });

  const columns: GridColDef[] = [
    { field: 'username', headerName: '用户名', width: 150 },
    { field: 'email', headerName: '邮箱', width: 200 },
    { field: 'role', headerName: '角色', width: 120 },
    {
      field: 'mfa_enabled',
      headerName: 'MFA 状态',
      width: 100,
      renderCell: (params) => (
        <span style={{ color: params.value ? 'green' : 'gray' }}>
          {params.value ? '已启用' : '未启用'}
        </span>
      ),
    },
    {
      field: 'actions',
      headerName: '操作',
      width: 150,
      renderCell: (params) => (
        <>
          <Button size="small" onClick={() => handleEdit(params.row.id)}>
            编辑
          </Button>
          <Button size="small" color="error" onClick={() => handleDelete(params.row.id)}>
            删除
          </Button>
        </>
      ),
    },
  ];

  const handleEdit = (id: string) => {
    // TODO: 打开编辑对话框
  };

  const handleDelete = async (id: string) => {
    if (confirm('确认删除该用户？')) {
      await userAPI.delete(id);
      queryClient.invalidateQueries({ queryKey: ['users'] });
    }
  };

  return (
    <div>
      <h1>用户管理</h1>
      <div style={{ marginBottom: 16 }}>
        <TextField
          label="搜索用户"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          size="small"
        />
        <Button variant="contained" onClick={() => setOpenDialog(true)} style={{ marginLeft: 16 }}>
          添加用户
        </Button>
      </div>

      <DataGrid
        rows={users?.data?.data || []}
        columns={columns}
        loading={isLoading}
        pagination
        paginationMode="server"
        rowCount={users?.data?.total || 0}
        page={page}
        onPageChange={(newPage) => setPage(newPage)}
        pageSize={20}
      />

      <Dialog open={openDialog} onClose={() => setOpenDialog(false)}>
        <DialogTitle>添加用户</DialogTitle>
        <DialogContent>
          {/* TODO: 添加用户表单 */}
        </DialogContent>
      </Dialog>
    </div>
  );
};
