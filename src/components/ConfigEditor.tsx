import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const onBaseUrlChange = (event: ChangeEvent<HTMLInputElement>) => {
    const jsonData = {
      ...options.jsonData,
      baseUrl: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  // Secure field (only sent to the backend)
  const onTokenChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        token: event.target.value,
      },
    });
  };

  const onResetToken = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        token: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        token: '',
      },
    });
  };

  const { jsonData, secureJsonFields } = options;
  const secureJsonData = (options.secureJsonData || {}) as MySecureJsonData;

  return (
    <div className="gf-form-group">
      <InlineField label="Base url" labelWidth={12}>
        <Input
          onChange={onBaseUrlChange}
          value={jsonData.baseUrl || ''}
          placeholder="https://sandbox.senx.io/api/v0"
          width={40}
        />
      </InlineField>
      <InlineField label="Read token" labelWidth={12}>
        <SecretInput
          isConfigured={(secureJsonFields && secureJsonFields.token) as boolean}
          value={secureJsonData.token || ''}
          placeholder="xxxx"
          width={40}
          onReset={onResetToken}
          onChange={onTokenChange}
        />
      </InlineField>
    </div>
  );
}
