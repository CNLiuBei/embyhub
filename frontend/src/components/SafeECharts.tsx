import React, { useEffect, useState, useRef, useCallback } from 'react';
import ReactECharts from 'echarts-for-react';

interface SafeEChartsProps {
  option: Record<string, any>;  // 使用宽松类型避免复杂的ECharts类型问题
  style?: React.CSSProperties;
}

/**
 * 安全的ECharts组件
 * 解决React 18 StrictMode下echarts-for-react的disconnect错误
 */
const SafeECharts: React.FC<SafeEChartsProps> = ({ option, style }) => {
  const [mounted, setMounted] = useState(false);
  const chartRef = useRef<any>(null);
  const isMountedRef = useRef(true);

  useEffect(() => {
    isMountedRef.current = true;
    // 延迟挂载，避免StrictMode双渲染问题
    const timer = setTimeout(() => {
      if (isMountedRef.current) {
        setMounted(true);
      }
    }, 10);
    
    return () => {
      isMountedRef.current = false;
      clearTimeout(timer);
      // 安全销毁图表
      try {
        if (chartRef.current) {
          const instance = chartRef.current.getEchartsInstance?.();
          if (instance && !instance.isDisposed?.()) {
            instance.dispose();
          }
        }
      } catch {
        // 忽略销毁错误
      }
    };
  }, []);

  const onChartReady = useCallback((instance: any) => {
    if (chartRef.current && isMountedRef.current) {
      chartRef.current = instance;
    }
  }, []);

  if (!mounted) {
    return <div style={{ ...style, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999' }}>加载中...</div>;
  }

  return (
    <ReactECharts
      ref={chartRef}
      option={option}
      style={style}
      notMerge={true}
      lazyUpdate={true}
      onChartReady={onChartReady}
    />
  );
};

export default SafeECharts;
