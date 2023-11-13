import http from 'k6/http'

export default function() {
  let baseUrl = 'http://34.124.148.20';
  let response = http.get(`${baseUrl}/products`);
  console.log(response.json());
}
