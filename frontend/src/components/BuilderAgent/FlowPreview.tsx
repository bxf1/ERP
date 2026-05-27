import React, { useMemo } from 'react';
import { Card, Tag, Typography, Empty } from 'antd';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import type { ApprovalFlowConfig, ApprovalNode } from '@/types';

const { Text } = Typography;

interface Props {
  flowConfig: ApprovalFlowConfig;
}

const nodeColors: Record<string, string> = {
  start: '#52c41a',
  approval: '#1677ff',
  condition: '#faad14',
  end: '#8c8c8c',
};

const nodeLabels: Record<string, string> = {
  start: '开始',
  approval: '审批',
  condition: '条件',
  end: '结束',
};

function buildFlowNodes(nodes: ApprovalNode[]) {
  return nodes.map((node, i) => ({
    id: node.id,
    position: { x: 250, y: i * 120 },
    data: {
      label: (
        <div style={{ textAlign: 'center', padding: '8px 16px' }}>
          <Tag color={nodeColors[node.type]}>{nodeLabels[node.type]}</Tag>
          <div style={{ fontWeight: 600, marginTop: 4 }}>{node.label}</div>
          {node.assignee && (
            <Text style={{ fontSize: 11, color: '#888' }}>
              负责人: {node.assignee}
            </Text>
          )}
        </div>
      ),
    },
    type: 'default' as const,
    style: {
      background: '#fff',
      border: `2px solid ${nodeColors[node.type]}`,
      borderRadius: 12,
      width: 180,
    },
  }));
}

function buildFlowEdges(edges: ApprovalFlowConfig['edges']) {
  return edges.map((e) => ({
    id: e.id,
    source: e.source,
    target: e.target,
    label: e.label || '',
    style: { stroke: '#bbb' },
    animated: true,
  }));
}

const FlowPreview: React.FC<Props> = ({ flowConfig }) => {
  const initialNodes = useMemo(
    () => buildFlowNodes(flowConfig.nodes),
    [flowConfig.nodes],
  );
  const initialEdges = useMemo(
    () => buildFlowEdges(flowConfig.edges),
    [flowConfig.edges],
  );

  const [nodes] = useNodesState(initialNodes);
  const [edges] = useEdgesState(initialEdges);

  if (!flowConfig.nodes.length) {
    return (
      <Card title="流程预览" size="small" style={{ marginBottom: 16 }}>
        <Empty description="暂无流程配置" />
      </Card>
    );
  }

  return (
    <Card title={`流程预览 — ${flowConfig.name}`} size="small" style={{ marginBottom: 16 }}>
      <div style={{ height: 400, border: '1px solid #f0f0f0', borderRadius: 8 }}>
        <ReactFlow nodes={nodes} edges={edges} fitView>
          <Background color="#f0f0f0" gap={16} />
          <Controls />
          <MiniMap
            nodeColor={(n) => {
              const type = flowConfig.nodes.find(
                (fn) => fn.id === n.id,
              )?.type;
              return type ? nodeColors[type] : '#ddd';
            }}
          />
        </ReactFlow>
      </div>
    </Card>
  );
};

export default FlowPreview;
