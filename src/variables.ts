import { Observable } from 'rxjs';
import { CustomVariableSupport, DataQueryRequest, DataQueryResponse } from '@grafana/data';
import { VariableQueryEditor } from './components/VariableQueryEditor';
import { assign } from 'lodash';
import { DataSource } from './datasource';
import { Warp10Query } from './types';

export class Warp10VariableSupport extends CustomVariableSupport<DataSource, Warp10Query> {
  constructor(private readonly datasource: DataSource) {
    super();
    this.datasource = datasource;
    this.query = this.query.bind(this);
  }

  editor = VariableQueryEditor;

  query(request: DataQueryRequest<Warp10Query>): Observable<DataQueryResponse> {
    assign(request.targets, [{ ...request.targets[0], refId: 'A' }]);

    return this.datasource.query(request);
  }
}
