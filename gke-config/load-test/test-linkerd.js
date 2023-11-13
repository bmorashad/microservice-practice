import http from 'k6/http'

export default function() {
  let baseUrl = 'http://35.244.161.62';
  let response = http.get(`${baseUrl}/products`);
  console.log(response.json());
}
