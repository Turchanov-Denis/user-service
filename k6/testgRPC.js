//$env:API_URL="your_ip:port"; k6 run --vus 10 --iterations 1000 testgRPC.js PowerShell
//set API_URL=your_ip:port && k6 run --vus 10 --iterations 1000 testgRPC.js CMD

import grpc from 'k6/net/grpc';
import { check } from 'k6';

const GRPC_URL = __ENV.GRPC_URL || '127.0.0.1:8080';

const client = new grpc.Client();
client.load([], '../user.proto');

export default () => {
    client.connect(GRPC_URL, {
        plaintext: true,
    });

    const id = Math.floor(Math.random() * 1000) + 1;

    const response = client.invoke(
        'user.UserService/GetUser',
        { id: `${id}` }
    );

    check(response, {
        'status is OK': (r) => r && r.status === grpc.StatusOK,
    });

    client.close();
}