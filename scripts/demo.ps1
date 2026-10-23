param([string]$BaseURL = 'http://localhost:8080', [string]$AdminKey = $env:API_KEY)
$ErrorActionPreference = 'Stop'
if (-not $AdminKey) {
    $envPath = Join-Path (Split-Path -Parent $PSScriptRoot) '.env'
    if (Test-Path -LiteralPath $envPath) {
        $line = Get-Content -LiteralPath $envPath | Where-Object { $_ -match '^API_KEY=' } | Select-Object -First 1
        if ($line) { $AdminKey = ($line -split '=', 2)[1].Trim() }
    }
}
if (-not $AdminKey) { throw 'Set API_KEY or pass -AdminKey. This script uses the credential only for administrator operations.' }
function Call-API($Method, $Path, $Body, $Headers = @{}) {
    Start-Sleep -Milliseconds 220
    $params = @{ Method = $Method; Uri = "$BaseURL$Path"; Headers = $Headers }
    if ($null -ne $Body) { $params.ContentType = 'application/json'; $params.Body = ($Body | ConvertTo-Json -Depth 8 -Compress) }
    return Invoke-RestMethod @params
}
$suffix = [guid]::NewGuid().ToString('N').Substring(0, 10)
$admin = @{ 'X-API-Key' = $AdminKey }
$phoneCategory = (Call-API POST '/api/v1/categories' @{ name = "Phones Demo $suffix"; slug = "phones-$suffix" } $admin).data
$laptopCategory = (Call-API POST '/api/v1/categories' @{ name = "Laptops Demo $suffix"; slug = "laptops-$suffix" } $admin).data
$phone = (Call-API POST '/api/v1/products' @{ category_id = $phoneCategory.id; sku = "PHONE-$suffix"; name = 'Demo Phone 128GB'; slug = "phone-$suffix"; brand = 'Demo'; model = 'Phone'; price_vnd = 15000000; stock_quantity = 10; specifications = @{ ram = '8GB'; storage = '128GB' } } $admin).data
$laptop = (Call-API POST '/api/v1/products' @{ category_id = $laptopCategory.id; sku = "LAPTOP-$suffix"; name = 'Demo Laptop 16GB'; slug = "laptop-$suffix"; brand = 'Demo'; model = 'Laptop'; price_vnd = 20000000; stock_quantity = 5; specifications = @{ ram = '16GB'; storage = '512GB SSD'; screen = '14 inch' } } $admin).data
$email = "demo-$suffix@example.com"
$password = 'Demo1!' + [guid]::NewGuid().ToString('N')
$null = Call-API POST '/api/v1/auth/register' @{ name = 'Demo Buyer'; email = $email; age = 21; password = $password }
$login = (Call-API POST '/api/v1/auth/login' @{ email = $email; password = $password }).data
$buyer = @{ Authorization = "Bearer $($login.access_token)" }
$null = Call-API PUT "/api/v1/cart/items/$($phone.id)" @{ quantity = 2 } $buyer
$checkout = @{ Authorization = $buyer.Authorization; 'Idempotency-Key' = [guid]::NewGuid().ToString() }
$address = @{ recipient_name = 'Demo Buyer'; phone = '0901234567'; address = 'Thanh Tri, Ha Noi' }
$order = (Call-API POST '/api/v1/orders' $address $checkout).data
$replay = (Call-API POST '/api/v1/orders' $address $checkout).data
if ($order.id -ne $replay.id -or $order.total_vnd -ne 30000000) { throw 'Checkout/idempotency validation failed' }
$null = Call-API POST "/api/v1/orders/$($order.id)/cancel" $null $buyer
$after = (Call-API GET "/api/v1/products/$($phone.id)" $null).data
if ($after.stock_quantity -ne 10) { throw 'Cancellation did not restore stock' }
Write-Output "MVP demo passed. Order: $($order.id); duplicate retry returned the same order; cancelled stock restored to 10."
Write-Output "Phone: $($phone.id); Laptop: $($laptop.id). Demo data remains available; no existing records were removed."
