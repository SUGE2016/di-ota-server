INSERT INTO t_role (role_code, display_name)
VALUES ('secret_admin', '设备 Secret 管理员')
ON CONFLICT (role_code) DO NOTHING;
