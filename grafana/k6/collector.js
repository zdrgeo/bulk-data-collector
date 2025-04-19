import http from 'k6/http';
import { check } from 'k6';
import { randomString, randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

const device = {
  oui: '766768',
  productClass: 'ONU',
};

const deviceSerialNumbers = [
  '400102030405',
  '400102030406',
  '400102030407',
  '400102030408',
  '400102030409',
  '400102030410',
  '400102030411',
  '400102030412',
  '400102030413',
  '400102030414',
];

function csvParameterPerRow() {
  // const deviceSerialNumber = '400102030405',
  const deviceSerialNumber = randomItem(deviceSerialNumbers);
  // const deviceSerialNumber = randomString(12, '0123456789ABCDEF')
  // const deviceSerialNumber = randomString(4, '01AB') // Bound to range of 256 serial numbers

  const url = `http://localhost:8088/collector?oui=${device.oui}&pc=${device.productClass}&sn=${deviceSerialNumber}`;

  const body = `ReportTimestamp,ParameterName,ParameterValue,ParameterType
1364529149,Device.MoCA.Interface.1.Stats.BroadPktSent,25248,unsignedLong
1364529149,Device.MoCA.Interface.1.Stats.BytesReceived,200543250,unsignedLong
1364529149,Device.MoCA.Interface.1.Stats.BytesSent,7682161,unsignedLong
1364529149,Device.MoCA.Interface.1.Stats.MultiPktReceived,890682272,unsignedLong`;

  const params = {
    headers: {
      'Content-Type': 'text/csv; charset=UTF-8; header=present',
      'BBF-Report-Format': 'ParameterPerRow',
    },
    tags: {
      name: 'CollectorURL',
    },
  };

  const res = http.post(url, body, params);

  check(res, {
    'is status 200': (r) => r.status === 200,
  });
}

function jsonNameValuePair() {
  // const deviceSerialNumber = '400102030405',
  const deviceSerialNumber = randomItem(deviceSerialNumbers);
  // const deviceSerialNumber = randomString(12, '0123456789ABCDEF')
  // const deviceSerialNumber = randomString(4, '01AB') // Bound to range of 265 serial numbers

  const url = `http://localhost:8088/collector?oui=${device.oui}&pc=${device.productClass}&sn=${deviceSerialNumber}`;

  const body = JSON.stringify({
    'Report': [
      {
        'CollectionTime': 1364529149,
        'Device.MoCA.Interface.1.Stats.BroadPktSent': 25248,
        'Device.MoCA.Interface.1.Stats.BytesReceived': 200543250,
        'Device.MoCA.Interface.1.Stats.BytesSent': 25248,
        'Device.MoCA.Interface.1.Stats.MultiPktReceived': 200543250,
        'Device.MoCA.Interface.2.Stats.BroadPktSent': 93247,
        'Device.MoCA.Interface.2.Stats.BytesReceived': 900543250,
        'Device.MoCA.Interface.2.Stats.BytesSent': 93247,
        'Device.MoCA.Interface.2.Stats.MultiPktReceived': 900543250
      }
    ]
  });

  const params = {
    headers: {
      'Content-Type': 'application/json; charset=UTF-8',
      'BBF-Report-Format': 'NameValuePair',
    },
    tags: {
      name: 'CollectorURL',
    },
  };

  const res = http.post(url, body, params);

  check(res, {
    'is status 200': (r) => r.status === 200,
  });
}

export const options = {
  vus: 1,
  duration: '300s',
};

export default csvParameterPerRow;
