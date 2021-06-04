import { defaults } from 'lodash';

import React, { PureComponent } from 'react';
import { CodeEditor } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from './datasource';
import { defaultQuery, MyDataSourceOptions, MyQuery } from './types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  onQueryTextChange = (value: string) => {
    const { onChange, query } = this.props;
    onChange({ ...query, queryText: value });
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { queryText } = query;

    return (
      <div className="gf-form">
        <CodeEditor
          value={queryText || ''}
          onBlur={this.onQueryTextChange}
          language="warpscript"
          width="95vw"
          height="20vh"
          showLineNumbers={true}
          showMiniMap={false}
        />
      </div>
    );
  }
}
