import React, { ChangeEvent } from 'react';
import { CodeEditor, InlineField, Input, InlineSwitch, VerticalGroup } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { Warp10DataSourceOptions, Warp10Query } from '../types';

type Props = QueryEditorProps<DataSource, Warp10Query, Warp10DataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onWarpscriptChange = (text: string) => {
    onChange({ ...query, warpscript: text });
  };

  const onAliasChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, alias: event.target.value });
  };

  const onShowLabelsChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, showLabels: event.target.checked });
  };

  const { warpscript, alias, showLabels } = query;

  const containerStyle = {
    width: "100%"
  }

  const horizontalContainer = {
    display: "flex",
    width: "100%"
  }

  return (
    <VerticalGroup>
      <div style={containerStyle}>
        <CodeEditor
          aria-label="Warpscript"
          value={warpscript || ''}
          onBlur={onWarpscriptChange}
          onSave={onWarpscriptChange}
          language="warpscript"
          width="100%"
          height="20vh"
          showLineNumbers={true}
          showMiniMap={false}
        />
      </div>
      <div style={horizontalContainer}>
        <InlineField label="Show labels">
          <InlineSwitch value={showLabels} onChange={onShowLabelsChange} />
        </InlineField>
        <InlineField
          label="Alias by"
          grow={true}
        >
          <Input
            onChange={onAliasChange}
            value={alias || ''}
            placeholder="Naming pattern"
          />
        </InlineField>
      </div>
    </VerticalGroup>
  );
}
