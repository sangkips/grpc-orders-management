package main

import (
	"context"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/sangkips/order-processing-system/services/common/genproto/orders/orders"
)

type httpServer struct {
	addr string
}

func NewHttpServer(addr string) *httpServer {
	return &httpServer{addr: addr}
}

func (s *httpServer) Run() error {
	router := http.NewServeMux()

	conn := NewGRPCClient(":9000")
	defer conn.Close()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c := orders.NewOrderServiceClient(conn)

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
		defer cancel()

		_, err := c.CreateOrder(ctx, &orders.CreateOrderRequest{
			CustomerID: 24,
			ProductID:  3123,
			Quantity:   2,
		})
		if err != nil {
			log.Printf("client error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		res, err := c.GetOrders(ctx, &orders.GetOrdersRequest{
			CustomerID: 42,
		})
		if err != nil {
			log.Printf("client error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		t := template.Must(template.New("orders").Parse(ordersTemplate))

		type TemplateData struct {
			Orders      []*orders.Order
			LastUpdated string
		}
		
		data := TemplateData{
			Orders:      res.GetOrders(),
			LastUpdated: time.Now().Format("Jan 02, 2025 15:04 MST"),
		}
		
		if err := t.Execute(w, data); err != nil {
			log.Printf("template error: %v", err)
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			return
		}
	})

	log.Println("Starting server on", s.addr)
	return http.ListenAndServe(s.addr, router)
}

var ordersTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Kitchen Orders</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f7fa;
        }
        
        h1 {
            color: #2c3e50;
            text-align: center;
            margin-bottom: 30px;
            font-size: 2.5rem;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
        }
        
        .filter-form {
            background-color: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            margin-bottom: 30px;
        }
        
        .form-group {
            display: flex;
            align-items: center;
            gap: 15px;
            flex-wrap: wrap;
        }
        
        label {
            font-weight: 600;
            min-width: 150px;
        }
        
        input[type="number"] {
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            flex-grow: 1;
            min-width: 100px;
        }
        
        button {
            background-color: #3498db;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 4px;
            cursor: pointer;
            font-weight: 600;
            transition: background-color 0.3s;
        }
        
        button:hover {
            background-color: #2980b9;
        }
        
        .orders-table {
            width: 100%;
            border-collapse: collapse;
            background-color: white;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            border-radius: 8px;
            overflow: hidden;
        }
        
        .orders-table th {
            background-color: #3498db;
            color: white;
            padding: 15px;
            text-align: left;
        }
        
        .orders-table td {
            padding: 12px 15px;
            border-bottom: 1px solid #eee;
        }
        
        .orders-table tr:last-child td {
            border-bottom: none;
        }
        
        .orders-table tr:nth-child(even) {
            background-color: #f8f9fa;
        }
        
        .orders-table tr:hover {
            background-color: #f1f4f7;
        }
        
        .empty-message {
            text-align: center;
            padding: 20px;
            font-style: italic;
            color: #7f8c8d;
        }
        
        .footer {
            margin-top: 30px;
            text-align: center;
            font-size: 0.9rem;
            color: #7f8c8d;
        }
    </style>
</head>
<body>
    <h1>Orders Management</h1>
    
    <form class="filter-form" action="/" method="get">
        <div class="form-group">
            <label for="customer_id">Filter by Customer ID:</label>
            <input type="number" id="customer_id" name="customer_id" placeholder="Enter customer ID">
            <button type="submit">Apply Filter</button>
        </div>
    </form>
    
    <table class="orders-table">
        <thead>
            <tr>
                <th>Order ID</th>
                <th>Customer ID</th>
                <th>Quantity</th>
                <th>Status</th>
            </tr>
        </thead>
        <tbody>
            {{if .Orders}}
                {{range .Orders}}
                <tr>
                    <td>#{{.OrderID}}</td>
                    <td>{{.CustomerID}}</td>
                    <td>{{.Quantity}}</td>
                    <td><span class="status-badge">Processing</span></td>
                </tr>
                {{end}}
            {{else}}
                <tr>
                    <td colspan="4" class="empty-message">No orders found. Try a different filter or add new orders.</td>
                </tr>
            {{end}}
        </tbody>
    </table>
    
    <div class="footer">
        <p>© 2025 Order Processing System | Last updated: {{.LastUpdated}}</p>
    </div>
</body>
</html>`