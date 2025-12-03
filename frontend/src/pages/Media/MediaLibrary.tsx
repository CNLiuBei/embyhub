import React, { useState, useEffect } from 'react';
import { Row, Col, Spin, Empty, Tabs, Select, Input, message, Modal, Space, Button } from 'antd';
import { 
  PlayCircleOutlined, 
  VideoCameraOutlined, 
  CustomerServiceOutlined,
  FolderOutlined,
  StarFilled,
  SearchOutlined,
  LeftOutlined,
  RightOutlined,
  AppstoreOutlined,
  UnorderedListOutlined,
  SortAscendingOutlined,
  SortDescendingOutlined
} from '@ant-design/icons';
import { getLibraries, getItems, getLatestItems, getServerUrl, MediaLibrary, MediaItem } from '@/api/emby';

// 媒体类型图标映射
const typeIcons: Record<string, React.ReactNode> = {
  movies: <VideoCameraOutlined />,
  tvshows: <PlayCircleOutlined />,
  music: <CustomerServiceOutlined />,
  default: <FolderOutlined />
};

// 精简媒体卡片组件
const MediaCard: React.FC<{ item: MediaItem; serverUrl: string; compact?: boolean }> = ({ item, serverUrl, compact }) => {
  const [imageError, setImageError] = useState(false);
  const [isHovered, setIsHovered] = useState(false);
  const imageUrl = item.ImageTags?.Primary && serverUrl && !imageError
    ? `${serverUrl}/Items/${item.Id}/Images/Primary?maxWidth=400&tag=${item.ImageTags.Primary}`
    : null;

  return (
    <div
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      style={{
        position: 'relative',
        borderRadius: 12,
        overflow: 'hidden',
        cursor: 'pointer',
        transform: isHovered ? 'scale(1.03)' : 'scale(1)',
        transition: 'all 0.3s ease',
        boxShadow: isHovered ? '0 12px 28px rgba(0,0,0,0.2)' : '0 4px 12px rgba(0,0,0,0.1)',
        background: '#1a1a2e',
      }}
    >
      {/* 封面图 - 使用2:3海报比例完整显示 */}
      <div style={{ 
        position: 'relative', 
        paddingBottom: compact ? '140%' : '150%',  // 2:3 海报比例
        background: '#1a1a2e'
      }}>
        {imageUrl ? (
          <img 
            src={imageUrl} 
            alt={item.Name}
            style={{ 
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%', 
              height: '100%', 
              objectFit: 'cover',  // 填满容器，无黑边
              display: 'block' 
            }}
            onError={() => setImageError(true)}
          />
        ) : (
          <div style={{ 
            position: 'absolute',
            top: 0,
            left: 0,
            width: '100%',
            height: '100%',
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}>
            <PlayCircleOutlined style={{ fontSize: 48, color: 'rgba(255,255,255,0.5)' }} />
          </div>
        )}
      </div>

      {/* 右上角评分 */}
      {item.CommunityRating && (
        <div style={{
          position: 'absolute',
          top: 8,
          right: 8,
          background: 'rgba(0,0,0,0.7)',
          borderRadius: 6,
          padding: '3px 6px',
          display: 'flex',
          alignItems: 'center',
          gap: 3,
        }}>
          <StarFilled style={{ fontSize: 10, color: '#ffc107' }} />
          <span style={{ color: '#fff', fontSize: 11, fontWeight: 600 }}>
            {item.CommunityRating.toFixed(1)}
          </span>
        </div>
      )}
      
      {/* 底部信息遮罩 */}
      <div style={{
        position: 'absolute',
        bottom: 0,
        left: 0,
        right: 0,
        padding: compact ? '20px 8px 8px' : '28px 10px 10px',
        background: 'linear-gradient(transparent, rgba(0,0,0,0.85))',
      }}>
        <div style={{ 
          color: '#fff', 
          fontWeight: 600, 
          fontSize: compact ? 11 : 13,
          overflow: 'hidden', 
          textOverflow: 'ellipsis', 
          whiteSpace: 'nowrap',
          textShadow: '0 1px 2px rgba(0,0,0,0.5)'
        }}>
          {item.Name}
        </div>
        {item.ProductionYear && (
          <div style={{ color: 'rgba(255,255,255,0.7)', fontSize: compact ? 10 : 11, marginTop: 2 }}>
            {item.ProductionYear}
          </div>
        )}
      </div>

    </div>
  );
};

const MediaLibraryPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [libraries, setLibraries] = useState<MediaLibrary[]>([]);
  const [items, setItems] = useState<MediaItem[]>([]);
  const [latestItems, setLatestItems] = useState<MediaItem[]>([]);
  const [total, setTotal] = useState(0);
  const [serverUrl, setServerUrl] = useState('');
  const [activeLibrary, setActiveLibrary] = useState<string>('');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [sortBy, setSortBy] = useState('SortName');
  const [sortOrder, setSortOrder] = useState('Ascending');
  const [page, setPage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  
  // 搜索弹窗相关状态
  const [searchModalVisible, setSearchModalVisible] = useState(false);
  const [searchResults, setSearchResults] = useState<MediaItem[]>([]);
  const [searchLoading, setSearchLoading] = useState(false);
  const [searchTotal, setSearchTotal] = useState(0);

  // 加载Emby服务器配置
  const loadServerConfig = async () => {
    try {
      const res = await getServerUrl();
      if (res.code === 200 && res.data) {
        setServerUrl(res.data.server_url);
      }
    } catch (error) {
      console.error('加载服务器配置失败');
    }
  };

  // 加载媒体库列表
  const loadLibraries = async () => {
    try {
      const res = await getLibraries();
      if (res.code === 200 && res.data) {
        setLibraries(res.data);
        // 默认选择第一个媒体库
        if (res.data.length > 0 && !activeLibrary) {
          const firstLib = res.data[0];
          setActiveLibrary(firstLib.Id || firstLib.ItemId || '');
        }
      }
    } catch (error) {
      message.error('加载媒体库失败');
    }
  };

  // 加载媒体项目
  const loadItems = async () => {
    if (!activeLibrary) return;
    setLoading(true);
    try {
      const res = await getItems({
        parent_id: activeLibrary,
        page,
        page_size: 48,
        sort_by: sortBy,
        sort_order: sortOrder
      });
      if (res.code === 200 && res.data) {
        setItems(res.data.list || []);
        setTotal(res.data.total || 0);
      }
    } catch (error) {
      message.error('加载媒体列表失败');
    } finally {
      setLoading(false);
    }
  };

  // 搜索媒体
  const doSearch = async (keyword: string) => {
    if (!keyword.trim()) return;
    setSearchLoading(true);
    setSearchModalVisible(true);
    try {
      const res = await getItems({
        page: 1,
        page_size: 50,
        sort_by: 'SortName',
        sort_order: 'Ascending',
        search: keyword
      });
      if (res.code === 200 && res.data) {
        setSearchResults(res.data.list || []);
        setSearchTotal(res.data.total || 0);
      }
    } catch (error) {
      message.error('搜索失败');
    } finally {
      setSearchLoading(false);
    }
  };

  // 加载最新媒体
  const loadLatest = async () => {
    try {
      const res = await getLatestItems({ limit: 12 });
      if (res.code === 200 && res.data) {
        setLatestItems(res.data);
      }
    } catch (error) {
      console.error('加载最新媒体失败');
    }
  };

  // 搜索处理（防抖）
  const handleSearch = (value: string) => {
    setSearchKeyword(value);
  };

  // 按回车或点击搜索
  const handleSearchSubmit = () => {
    if (searchKeyword.trim()) {
      doSearch(searchKeyword);
    }
  };

  useEffect(() => {
    loadServerConfig();
    loadLibraries();
    loadLatest();
  }, []);

  useEffect(() => {
    if (activeLibrary) {
      loadItems();
    }
  }, [activeLibrary, page, sortBy, sortOrder]);

  // 获取媒体库图标
  const getLibraryIcon = (type: string) => {
    return typeIcons[type] || typeIcons.default;
  };

  return (
    <div style={{ padding: '0 4px' }}>
      {/* 页面头部 */}
      <div style={{ 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center', 
        marginBottom: 28,
        flexWrap: 'wrap',
        gap: 16
      }}>
        <div>
          <h1 style={{ 
            fontSize: 28, 
            fontWeight: 700, 
            color: '#1d1d1f', 
            margin: 0,
            letterSpacing: '-0.5px'
          }}>
            媒体库
          </h1>
          <p style={{ color: '#86868b', marginTop: 4, fontSize: 14, margin: '4px 0 0' }}>
            浏览Emby服务器上的媒体内容
          </p>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <Space.Compact>
            <Input
              placeholder="搜索媒体..."
              style={{ width: 180 }}
              value={searchKeyword}
              onChange={(e) => handleSearch(e.target.value)}
              onPressEnter={handleSearchSubmit}
              allowClear
            />
            <Button type="primary" onClick={handleSearchSubmit}>搜索</Button>
          </Space.Compact>
          <Select
            value={sortBy}
            onChange={setSortBy}
            style={{ width: 120 }}
            options={[
              { value: 'SortName', label: '名称' },
              { value: 'DateCreated', label: '添加日期' },
              { value: 'PremiereDate', label: '首播日期' },
              { value: 'CommunityRating', label: '社区评分' },
              { value: 'CriticRating', label: '评论家评分' },
              { value: 'Runtime', label: '时长' },
              { value: 'Random', label: '随机' },
            ]}
          />
          <div
            onClick={() => setSortOrder(sortOrder === 'Ascending' ? 'Descending' : 'Ascending')}
            style={{
              padding: '7px 10px',
              borderRadius: 8,
              cursor: 'pointer',
              background: '#f5f5f7',
              display: 'flex',
              alignItems: 'center',
              transition: 'all 0.2s',
            }}
            title={sortOrder === 'Ascending' ? '升序' : '降序'}
          >
            {sortOrder === 'Ascending' ? (
              <SortAscendingOutlined style={{ fontSize: 16, color: '#007aff' }} />
            ) : (
              <SortDescendingOutlined style={{ fontSize: 16, color: '#007aff' }} />
            )}
          </div>
          <div style={{ display: 'flex', background: '#f5f5f7', borderRadius: 10, padding: 3 }}>
            <div
              onClick={() => setViewMode('grid')}
              style={{
                padding: '7px 12px',
                borderRadius: 8,
                cursor: 'pointer',
                background: viewMode === 'grid' ? '#fff' : 'transparent',
                boxShadow: viewMode === 'grid' ? '0 1px 3px rgba(0,0,0,0.1)' : 'none',
                transition: 'all 0.2s'
              }}
            >
              <AppstoreOutlined style={{ color: viewMode === 'grid' ? '#007aff' : '#86868b' }} />
            </div>
            <div
              onClick={() => setViewMode('list')}
              style={{
                padding: '7px 12px',
                borderRadius: 8,
                cursor: 'pointer',
                background: viewMode === 'list' ? '#fff' : 'transparent',
                boxShadow: viewMode === 'list' ? '0 1px 3px rgba(0,0,0,0.1)' : 'none',
                transition: 'all 0.2s'
              }}
            >
              <UnorderedListOutlined style={{ color: viewMode === 'list' ? '#007aff' : '#86868b' }} />
            </div>
          </div>
        </div>
      </div>

      {/* 最新添加 */}
      {latestItems.length > 0 && (
        <div style={{ 
          marginBottom: 32, 
          background: 'rgba(255, 255, 255, 0.5)',
          backdropFilter: 'blur(20px) saturate(180%)',
          borderRadius: 16,
          padding: '20px 0',
          boxShadow: '0 4px 20px rgba(0,0,0,0.08)',
          border: '1px solid rgba(255, 255, 255, 0.4)'
        }}>
          <div style={{ 
            display: 'flex', 
            justifyContent: 'space-between', 
            alignItems: 'center',
            marginBottom: 16,
            padding: '0 20px'
          }}>
            <h3 style={{ margin: 0, fontWeight: 600, color: '#1d1d1f', fontSize: 16 }}>
              🎬 最新添加
            </h3>
            <div style={{ display: 'flex', gap: 8 }}>
              <div
                onClick={() => {
                  const container = document.getElementById('latest-scroll');
                  if (container) container.scrollBy({ left: -300, behavior: 'smooth' });
                }}
                style={{
                  width: 32,
                  height: 32,
                  borderRadius: '50%',
                  background: '#f5f5f7',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  cursor: 'pointer',
                  transition: 'background 0.2s',
                }}
                onMouseEnter={(e) => e.currentTarget.style.background = '#e8e8ed'}
                onMouseLeave={(e) => e.currentTarget.style.background = '#f5f5f7'}
              >
                <LeftOutlined style={{ color: '#86868b', fontSize: 12 }} />
              </div>
              <div
                onClick={() => {
                  const container = document.getElementById('latest-scroll');
                  if (container) container.scrollBy({ left: 300, behavior: 'smooth' });
                }}
                style={{
                  width: 32,
                  height: 32,
                  borderRadius: '50%',
                  background: '#f5f5f7',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  cursor: 'pointer',
                  transition: 'background 0.2s',
                }}
                onMouseEnter={(e) => e.currentTarget.style.background = '#e8e8ed'}
                onMouseLeave={(e) => e.currentTarget.style.background = '#f5f5f7'}
              >
                <RightOutlined style={{ color: '#86868b', fontSize: 12 }} />
              </div>
            </div>
          </div>
          <div 
            id="latest-scroll"
            style={{ 
              display: 'flex', 
              gap: 12, 
              overflowX: 'auto', 
              padding: '0 20px 8px',
              scrollbarWidth: 'none',
              msOverflowStyle: 'none',
            }}
          >
            {latestItems.map((item) => (
              <div key={item.Id} style={{ minWidth: 110, maxWidth: 110, flexShrink: 0 }}>
                <MediaCard item={item} serverUrl={serverUrl} compact />
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 媒体库标签页 */}
      <div style={{ 
        background: 'rgba(255, 255, 255, 0.5)', 
        backdropFilter: 'blur(20px) saturate(180%)',
        borderRadius: 12, 
        padding: '4px 16px',
        marginBottom: 24,
        border: '1px solid rgba(255, 255, 255, 0.4)',
        boxShadow: '0 2px 8px rgba(0,0,0,0.06)'
      }}>
        <Tabs
          activeKey={activeLibrary}
          onChange={(key) => { setActiveLibrary(key); setPage(1); }}
          items={libraries.map(lib => ({
            key: lib.Id || lib.ItemId || lib.Name,
            label: (
              <span style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '4px 0' }}>
                {getLibraryIcon(lib.CollectionType)}
                <span style={{ fontWeight: 500 }}>{lib.Name}</span>
              </span>
            )
          }))}
        />
      </div>

      {/* 媒体数量统计 */}
      {total > 0 && (
        <div style={{ marginBottom: 16, color: '#86868b', fontSize: 13 }}>
          共 <span style={{ color: '#1d1d1f', fontWeight: 600 }}>{total}</span> 个项目
        </div>
      )}

      {/* 媒体列表 */}
      <Spin spinning={loading}>
        {items.length > 0 ? (
          viewMode === 'grid' ? (
            // 网格视图
            <Row gutter={[16, 16]}>
              {items.map((item) => (
                <Col key={item.Id} xs={6} sm={4} md={3} lg={2} xl={2}>
                  <MediaCard item={item} serverUrl={serverUrl} />
                </Col>
              ))}
            </Row>
          ) : (
            // 列表视图
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              {items.map((item) => {
                const imageUrl = item.ImageTags?.Primary && serverUrl
                  ? `${serverUrl}/Items/${item.Id}/Images/Primary?maxWidth=120&tag=${item.ImageTags.Primary}`
                  : null;
                return (
                  <div
                    key={item.Id}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 16,
                      padding: 12,
                      background: 'rgba(255, 255, 255, 0.5)',
                      backdropFilter: 'blur(16px) saturate(150%)',
                      borderRadius: 12,
                      boxShadow: '0 4px 16px rgba(0,0,0,0.08)',
                      border: '1px solid rgba(255, 255, 255, 0.4)',
                      cursor: 'pointer',
                      transition: 'all 0.2s',
                    }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.boxShadow = '0 4px 16px rgba(0,0,0,0.1)';
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.boxShadow = '0 2px 8px rgba(0,0,0,0.06)';
                    }}
                  >
                    {/* 封面 */}
                    <div style={{ 
                      width: 60, 
                      height: 90, 
                      borderRadius: 8, 
                      overflow: 'hidden',
                      background: '#1a1a2e',
                      flexShrink: 0
                    }}>
                      {imageUrl ? (
                        <img src={imageUrl} alt={item.Name} style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                      ) : (
                        <div style={{ width: '100%', height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                          <PlayCircleOutlined style={{ color: 'rgba(255,255,255,0.5)', fontSize: 24 }} />
                        </div>
                      )}
                    </div>
                    {/* 信息 */}
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div style={{ fontWeight: 600, fontSize: 15, color: '#1d1d1f', marginBottom: 4, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {item.Name}
                      </div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 12, color: '#86868b', fontSize: 13 }}>
                        {item.ProductionYear && <span>{item.ProductionYear}</span>}
                        {item.CommunityRating && (
                          <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                            <StarFilled style={{ color: '#ffc107', fontSize: 12 }} />
                            {item.CommunityRating.toFixed(1)}
                          </span>
                        )}
                        {item.Genres && item.Genres.length > 0 && (
                          <span>{item.Genres.slice(0, 2).join(' / ')}</span>
                        )}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )
        ) : (
          <Empty 
            description="暂无媒体内容" 
            style={{ 
              marginTop: 80, 
              padding: 40,
              background: '#fafafa',
              borderRadius: 16
            }} 
          />
        )}
      </Spin>

      {/* 加载更多 */}
      {items.length > 0 && items.length < total && (
        <div style={{ textAlign: 'center', marginTop: 32, marginBottom: 16 }}>
          <div 
            onClick={() => setPage(page + 1)}
            style={{ 
              display: 'inline-flex',
              alignItems: 'center',
              gap: 8,
              padding: '12px 32px',
              background: 'linear-gradient(135deg, #007aff 0%, #5856d6 100%)',
              color: '#fff',
              borderRadius: 25,
              cursor: 'pointer',
              fontWeight: 500,
              fontSize: 14,
              boxShadow: '0 4px 12px rgba(0,122,255,0.3)',
              transition: 'all 0.3s ease',
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'translateY(-2px)';
              e.currentTarget.style.boxShadow = '0 6px 20px rgba(0,122,255,0.4)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'translateY(0)';
              e.currentTarget.style.boxShadow = '0 4px 12px rgba(0,122,255,0.3)';
            }}
          >
            加载更多
          </div>
          <div style={{ marginTop: 8, color: '#86868b', fontSize: 12 }}>
            已加载 {items.length} / {total}
          </div>
        </div>
      )}

      {/* 搜索结果弹窗 */}
      <Modal
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <SearchOutlined style={{ color: '#007aff' }} />
            <span>搜索结果</span>
            {searchTotal > 0 && (
              <span style={{ color: '#86868b', fontSize: 14, fontWeight: 400 }}>
                （共 {searchTotal} 个）
              </span>
            )}
          </div>
        }
        open={searchModalVisible}
        onCancel={() => {
          setSearchModalVisible(false);
          setSearchKeyword('');
        }}
        footer={null}
        width={800}
        styles={{ body: { maxHeight: '70vh', overflowY: 'auto', padding: '16px 24px' } }}
      >
        <Spin spinning={searchLoading}>
          {searchResults.length > 0 ? (
            <Row gutter={[12, 12]}>
              {searchResults.map((item) => (
                <Col key={item.Id} xs={8} sm={6} md={4}>
                  <MediaCard item={item} serverUrl={serverUrl} compact />
                </Col>
              ))}
            </Row>
          ) : (
            <Empty description={searchLoading ? '搜索中...' : '未找到相关内容'} />
          )}
        </Spin>
      </Modal>
    </div>
  );
};

export default MediaLibraryPage;
