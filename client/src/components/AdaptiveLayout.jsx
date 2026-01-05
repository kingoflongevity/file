import React from 'react';
import { ConfigProvider, theme } from 'antd';

const AdaptiveLayout = ({ children, className = '', style = {} }) => {
  const { token } = theme.useToken();

  return (
    <ConfigProvider
      theme={{
        token: {
          borderRadius: 8,
        },
      }}
    >
      <div
        className={`adaptive-layout ${className}`}
        style={{
          width: '100%',
          maxWidth: '1200px',
          margin: '0 auto',
          padding: '0 16px',
          transition: 'all 0.3s ease-in-out',
          ...style,
        }}
      >
        {children}
      </div>
    </ConfigProvider>
  );
};

// 自适应卡片组件
export const AdaptiveCard = ({ children, className = '', style = {} }) => {
  return (
    <div
      className={`adaptive-card ${className}`}
      style={{
        width: '100%',
        marginBottom: '16px',
        transition: 'all 0.3s ease-in-out',
        ...style,
      }}
    >
      {children}
    </div>
  );
};

// 自适应按钮组组件
export const AdaptiveButtonGroup = ({ children, className = '', style = {} }) => {
  return (
    <div
      className={`adaptive-button-group ${className}`}
      style={{
        display: 'flex',
        flexWrap: 'wrap',
        gap: '8px',
        marginBottom: '16px',
        ...style,
      }}
    >
      {children}
    </div>
  );
};

// 自适应网格组件
export const AdaptiveGrid = ({ children, columns = 3, className = '', style = {} }) => {
  return (
    <div
      className={`adaptive-grid ${className}`}
      style={{
        display: 'grid',
        gridTemplateColumns: `repeat(auto-fit, minmax(calc((100% - ${columns - 1} * 16px) / ${columns}), 1fr))`,
        gap: '16px',
        transition: 'all 0.3s ease-in-out',
        ...style,
      }}
    >
      {children}
    </div>
  );
};

export default AdaptiveLayout;