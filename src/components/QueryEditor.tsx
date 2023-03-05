import React from 'react';
import { CodeEditor } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onWarpscriptChange = (text: string) => {
    onChange({ ...query, warpscript: text });
  };

  const { warpscript } = query;

  const divStyle = {
    width: "100%"
  }

  return (
    <div style={divStyle}>
        <CodeEditor
          aria-label="Warpscript"
          value={warpscript || ''}
          onBlur={onWarpscriptChange}
          onSave={onWarpscriptChange}
          language="warpscript"
          height="20vh"
          showLineNumbers={true}
          showMiniMap={false}
        />
    </div>
  );
}
