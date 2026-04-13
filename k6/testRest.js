
//$env:API_URL="http://your_ip:port"; k6 run --vus 10 --iterations 1000 testRest.js PowerShell
//set API_URL=http://your_ip:port && k6 run --vus 10 --iterations 1000 testRest.js CMD

import http from 'k6/http';

const BASE_URL = __ENV.API_URL || 'http://127.0.0.1:8080';

export default function () {
    let id = Math.floor(Math.random() * 1000) + 1;
    http.get(`${BASE_URL}/user/${id}`);
}


// import http from 'k6/http';
//
// export default function () {
//     let id = Math.floor(Math.random() * 1000) + 1;
//
//     http.get(`http://127.0.0.1:8080/user/${id}`, {
//         headers: {
//             "Accept-Encoding": "gzip",
//         },
//     });
// }