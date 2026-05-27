import React from 'react';
import {
  Modal,
  Steps,
  Typography,
  Result,
  Button,
  Space,
  Alert,
  Progress,
} from 'antd';
import {
  CheckCircleFilled,
  CloseCircleFilled,
  LoadingOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons';
import type { CreationStepInfo, CreationResult } from '@/types';

const { Text, Paragraph } = Typography;

interface Props {
  open: boolean;
  progress: CreationStepInfo[] | null;
  result: CreationResult | null;
  onClose: () => void;
  onViewModule?: () => void;
}

const stepIcons: Record<string, React.ReactNode> = {
  waiting: <ClockCircleOutlined />,
  running: <LoadingOutlined />,
  done: <CheckCircleFilled style={{ color: '#52c41a' }} />,
  failed: <CloseCircleFilled style={{ color: '#ff4d4f' }} />,
};

const ProgressFeedback: React.FC<Props> = ({
  open,
  progress,
  result,
  onClose,
  onViewModule,
}) => {
  const getCurrentStep = () => {
    if (!progress) return 0;
    const runningIdx = progress.findIndex((s) => s.status === 'running');
    if (runningIdx >= 0) return runningIdx;
    const failedIdx = progress.findIndex((s) => s.status === 'failed');
    if (failedIdx >= 0) return failedIdx;
    return progress.length;
  };

  const hasFailed = progress?.some((s) => s.status === 'failed');
  const isDone = result?.success || progress?.every((s) => s.status === 'done');
  const percent = progress
    ? Math.round(
        (progress.filter((s) => s.status === 'done').length / progress.length) *
          100,
      )
    : 0;

  return (
    <Modal
      title={isDone ? '创建完成' : hasFailed ? '创建失败' : '正在创建模块…'}
      open={open}
      footer={null}
      closable={isDone || hasFailed}
      onCancel={onClose}
      width={540}
    >
      {/* Progress bar */}
      {progress && !isDone && !hasFailed && (
        <>
          <Progress
            percent={percent}
            status="active"
            style={{ marginBottom: 20 }}
          />
          <Steps
            direction="vertical"
            size="small"
            current={getCurrentStep()}
            items={progress.map((step) => ({
              title: step.label,
              description: step.message,
              icon: stepIcons[step.status],
              status:
                step.status === 'failed'
                  ? 'error'
                  : step.status === 'running'
                    ? 'process'
                    : step.status === 'done'
                      ? 'finish'
                      : 'wait',
            }))}
          />
        </>
      )}

      {/* Success */}
      {isDone && result?.success && (
        <Result
          status="success"
          title="模块创建成功！"
          subTitle={
            <span>
              模块 <Text strong>{result.moduleName}</Text> 已创建完成
              {result.moduleId && (
                <>
                  ，ID: <Text code>{result.moduleId}</Text>
                </>
              )}
            </span>
          }
          extra={
            <Space>
              {onViewModule && (
                <Button type="primary" onClick={onViewModule}>
                  查看模块
                </Button>
              )}
              <Button onClick={onClose}>关闭</Button>
            </Space>
          }
        >
          {result.links && (
            <div style={{ textAlign: 'left', marginTop: 16 }}>
              {result.links.formPage && (
                <Paragraph copyable={{ text: result.links.formPage }}>
                  表单页面: {result.links.formPage}
                </Paragraph>
              )}
              {result.links.flowPage && (
                <Paragraph copyable={{ text: result.links.flowPage }}>
                  流程页面: {result.links.flowPage}
                </Paragraph>
              )}
              {result.links.menuEntry && (
                <Paragraph copyable={{ text: result.links.menuEntry }}>
                  菜单入口: {result.links.menuEntry}
                </Paragraph>
              )}
            </div>
          )}
        </Result>
      )}

      {/* Failure */}
      {hasFailed && (
        <Result
          status="error"
          title="创建失败"
          subTitle="部分步骤执行失败，请查看详情后重试"
        >
          <Alert
            type="error"
            message="错误详情"
            description={
              <ul style={{ margin: 0, paddingLeft: 20 }}>
                {progress
                  ?.filter((s) => s.status === 'failed')
                  .map((s) => (
                    <li key={s.key}>
                      {s.label}: {s.message}
                    </li>
                  ))}
                {result?.errors?.map((e, i) => <li key={i}>{e}</li>)}
              </ul>
            }
            style={{ textAlign: 'left', marginBottom: 16 }}
          />
          <Button type="primary" onClick={onClose}>
            关闭
          </Button>
        </Result>
      )}
    </Modal>
  );
};

export default ProgressFeedback;
