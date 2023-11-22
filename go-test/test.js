import http from 'k6/http'


export default function() {
  let r = http.get(`https://jsonplaceholder.typicode.com/posts/1`);
  // console.log(r.json());
  // let response1 = http.get(`${baseUrl}/create-products/random`);
  // let response2 = http.get(`${baseUrl}/products`);
  // console.log(response1.json());
  // console.log(response2.json());
}
