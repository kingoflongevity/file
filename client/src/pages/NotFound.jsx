import React from 'react';
import { Link } from 'react-router-dom';

const NotFound = () => {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100 px-4">
      <div className="text-center">
        <h1 className="text-9xl font-bold text-gray-300">404</h1>
        <h2 className="text-3xl font-bold text-gray-800 mt-4">页面未找到</h2>
        <p className="text-gray-600 mt-2">抱歉，您请求的页面不存在或已被移除</p>
        <div className="mt-8">
          <Link
            to="/"
            className="px-6 py-3 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 transition-colors"
          >
            返回首页
          </Link>
        </div>
      </div>
    </div>
  );
};

export default NotFound;
