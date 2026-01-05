// 检查当前登录用户信息
const user = JSON.parse(localStorage.getItem('user'));
console.log('当前登录用户信息:', user);
console.log('用户角色:', user?.role);

// 检查是否为管理员
const isAdmin = user?.role === 'admin';
console.log('是否为管理员:', isAdmin);

// 检查localStorage中的token
const token = localStorage.getItem('token');
console.log('Token:', token);