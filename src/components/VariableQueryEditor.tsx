import React, { ChangeEvent } from 'react';
import { InlineField, Input, VerticalGroup } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { Warp10DataSourceOptions, Warp10Query } from '../types';

type Props = QueryEditorProps<DataSource, Warp10Query, Warp10DataSourceOptions>;

export function VariableQueryEditor({ query, onChange, onRunQuery }: Props) {
  const onClassChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, className: event.target.value });
  };

  const onLabelChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, labelName: event.target.value });
  };

  const { className, labelName } = query;

  return (
    <VerticalGroup>
      <div>
        <InlineField label="Class name" labelWidth={12}>
          <Input
            onChange={onClassChange}
            value={className || ''}
            width={40}
          />
        </InlineField>
        <InlineField label="Label name" labelWidth={12}>
          <Input
            onChange={onLabelChange}
            value={labelName || ''}
            width={40}
          />
        </InlineField>
      </div>
    </VerticalGroup>
  );
}
