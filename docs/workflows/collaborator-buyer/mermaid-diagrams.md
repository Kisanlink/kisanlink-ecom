# Collaborator-Buyer Mermaid Workflow Diagrams

## Workflow 1: Product Discovery & Browsing

```mermaid
flowchart TD
    A[User Login] --> B{Authentication Valid?}
    B -->|No| C[Authentication Error]
    B -->|Yes| D[Access Marketplace]

    D --> E[Browse Product Catalog]
    E --> F{Filter Applied?}
    F -->|Yes| G[Apply Filters<br/>Category, Price, Location]
    F -->|No| H[View All Products]
    G --> H

    H --> I[Select Product]
    I --> J[View Product Details]
    J --> K{Check Active Listings?}
    K -->|Yes| L[Search Active Listings]
    K -->|No| M[Continue Browsing]

    L --> N[View Listing Details]
    N --> O{Interested in Bidding?}
    O -->|Yes| P[Proceed to Bidding]
    O -->|No| Q[Continue Search]

    M --> E
    Q --> L
    C --> A

    style A fill:#e1f5fe
    style P fill:#c8e6c9
    style C fill:#ffcdd2
```

## Workflow 2: Auction Participation & Bidding

```mermaid
flowchart TD
    A[Select Listing to Bid] --> B{Pre-Bid Validation}
    B -->|Failed| C[Show Error Message<br/>- Expired listing<br/>- Own listing<br/>- Below minimum]
    B -->|Passed| D[Enter Bid Details]

    D --> E[Specify Bid Amount]
    E --> F[Set Quantity]
    F --> G[Choose Payment Method]
    G --> H[Add Message/Notes]
    H --> I{Configure Auto-Bid?}

    I -->|Yes| J[Set Auto-Bid Limit]
    I -->|No| K[Submit Bid]
    J --> K

    K --> L{Bid Validation}
    L -->|Failed| M[Validation Error<br/>- Amount too low<br/>- Insufficient funds<br/>- Invalid quantity]
    L -->|Passed| N[Process Bid]

    N --> O[Update Listing State]
    O --> P{Is Highest Bid?}
    P -->|Yes| Q[Become Highest Bidder]
    P -->|No| R[Bid Placed Successfully]

    Q --> S[Notify Previous Highest Bidder]
    S --> T[Send Confirmation to Bidder]
    R --> T

    T --> U{Monitor Auction}
    U --> V{Outbid?}
    V -->|Yes| W{Auto-Bid Enabled?}
    W -->|Yes| X{Within Auto-Bid Limit?}
    X -->|Yes| Y[Place Auto-Bid]
    X -->|No| Z[Notify Outbid]
    W -->|No| Z
    V -->|No| AA[Continue Monitoring]

    Y --> N
    Z --> BB{Place New Bid?}
    BB -->|Yes| D
    BB -->|No| CC[End Participation]
    AA --> U

    C --> A
    M --> D

    style A fill:#e1f5fe
    style Q fill:#c8e6c9
    style C fill:#ffcdd2
    style M fill:#ffcdd2
    style Z fill:#fff3e0
```

## Workflow 3: Auto-Bidding Management

```mermaid
flowchart TD
    A[Configure Auto-Bidding] --> B[Set Maximum Bid Limit]
    B --> C[Define Increment Strategy]
    C --> D[Set Response Timing]
    D --> E[Configure Notifications]
    E --> F[Activate Auto-Bidding]

    F --> G[Monitor for Competing Bids]
    G --> H{New Higher Bid Detected?}
    H -->|No| G
    H -->|Yes| I{Within Auto-Bid Limit?}

    I -->|No| J[Send Limit Reached Notification]
    I -->|Yes| K[Calculate Counter-Bid]

    K --> L{Counter-Bid Valid?}
    L -->|No| M[Log Auto-Bid Error]
    L -->|Yes| N[Place Auto-Bid]

    N --> O[Update Auction State]
    O --> P{Successfully Outbid Competitor?}
    P -->|Yes| Q[Send Success Notification]
    P -->|No| R[Log Auto-Bid Failure]

    Q --> S[Update Auto-Bid Status]
    R --> S
    J --> T[Disable Auto-Bidding]
    M --> T
    S --> G

    T --> U{Manual Intervention Needed?}
    U -->|Yes| V[Notify User for Action]
    U -->|No| W[Continue Monitoring]

    V --> X{User Updates Limits?}
    X -->|Yes| A
    X -->|No| Y[End Auto-Bidding]
    W --> G

    style A fill:#e1f5fe
    style F fill:#c8e6c9
    style J fill:#fff3e0
    style M fill:#ffcdd2
    style T fill:#ffcdd2
```

## Workflow 4: Bid History Tracking

```mermaid
flowchart TD
    A[Access Bid History] --> B{Select View Type}
    B -->|All Bids| C[Load Complete Bid History]
    B -->|Active Bids| D[Load Active Auctions Only]
    B -->|Won Bids| E[Load Winning Bids]
    B -->|Lost Bids| F[Load Unsuccessful Bids]

    C --> G[Apply Filters & Sorting]
    D --> G
    E --> G
    F --> G

    G --> H{Filter Options}
    H -->|Date Range| I[Filter by Date]
    H -->|Amount Range| J[Filter by Bid Amount]
    H -->|Product Category| K[Filter by Category]
    H -->|Status| L[Filter by Bid Status]

    I --> M[Display Filtered Results]
    J --> M
    K --> M
    L --> M

    M --> N{Select Specific Bid?}
    N -->|Yes| O[View Bid Details]
    N -->|No| P{Export Data?}

    O --> Q[Show Bid Information<br/>- Auction details<br/>- Bid progression<br/>- Competition analysis<br/>- Outcome status]

    P -->|Yes| R[Generate Export File<br/>- CSV format<br/>- PDF report<br/>- Excel spreadsheet]
    P -->|No| S[Continue Browsing]

    Q --> T{Analyze Performance?}
    T -->|Yes| U[Generate Analytics<br/>- Win/Loss ratio<br/>- Average bid amounts<br/>- Spending trends<br/>- Success patterns]
    T -->|No| V[Return to History]

    R --> W[Download File]
    S --> G
    U --> X[Display Analytics Dashboard]
    V --> M
    W --> Y[End Session]
    X --> Z{Export Analytics?}
    Z -->|Yes| R
    Z -->|No| V

    style A fill:#e1f5fe
    style U fill:#c8e6c9
    style X fill:#c8e6c9
    style W fill:#c8e6c9
```

## Workflow 5: Order Creation from Winning Bids

```mermaid
flowchart TD
    A[Auction Completed] --> B{Won Auction?}
    B -->|No| C[End Process]
    B -->|Yes| D[Receive Win Notification]

    D --> E[Access Order Creation]
    E --> F[Review Winning Bid Details<br/>- Product information<br/>- Final bid amount<br/>- Quantity<br/>- Seller details]

    F --> G[Enter Shipping Address]
    G --> H{Address Valid?}
    H -->|No| I[Address Validation Error]
    H -->|Yes| J[Select Payment Method]

    I --> G
    J --> K{Payment Method Valid?}
    K -->|No| L[Payment Method Error]
    K -->|Yes| M[Add Delivery Preferences]

    L --> J
    M --> N[Add Special Instructions]
    N --> O[Review Order Summary<br/>- Product details<br/>- Quantities and pricing<br/>- Total amount<br/>- Delivery info]

    O --> P{Confirm Order?}
    P -->|No| Q[Modify Order Details]
    P -->|Yes| R[Submit Order]

    Q --> G
    R --> S{Order Processing}
    S -->|Failed| T[Order Creation Error<br/>- Payment verification failed<br/>- Inventory not available<br/>- System error]
    S -->|Success| U[Order Created Successfully]

    T --> V{Retry Order?}
    V -->|Yes| R
    V -->|No| W[Cancel Order Creation]

    U --> X[Receive Order Confirmation<br/>- Order ID<br/>- Payment instructions<br/>- Delivery timeline<br/>- Seller contact info]

    X --> Y[Notify Seller of New Order]
    Y --> Z[Initiate Payment Process]
    Z --> AA[Track Order Status<br/>- Payment pending<br/>- Payment confirmed<br/>- In preparation<br/>- Shipped<br/>- Delivered]

    W --> C
    AA --> BB[Order Fulfillment Process]

    style A fill:#e1f5fe
    style D fill:#fff3e0
    style U fill:#c8e6c9
    style X fill:#c8e6c9
    style T fill:#ffcdd2
    style I fill:#ffcdd2
    style L fill:#ffcdd2
```

## Workflow 6: Payment Processing

```mermaid
flowchart TD
    A[Order Created] --> B[Select Payment Method]
    B --> C{Payment Method Type}
    C -->|Bank Transfer| D[Enter Bank Details]
    C -->|Digital Wallet| E[Select Wallet Provider]
    C -->|Credit Card| F[Enter Card Details]
    C -->|UPI| G[Enter UPI ID]

    D --> H[Verify Bank Account]
    E --> I[Authenticate Wallet]
    F --> J[Validate Card Information]
    G --> K[Verify UPI Handle]

    H --> L{Verification Successful?}
    I --> L
    J --> L
    K --> L

    L -->|No| M[Payment Method Error<br/>- Invalid details<br/>- Account not verified<br/>- Insufficient funds]
    L -->|Yes| N[Display Payment Summary<br/>- Amount breakdown<br/>- Payment method<br/>- Processing fees<br/>- Total amount]

    M --> B
    N --> O{Confirm Payment?}
    O -->|No| P[Cancel Payment]
    O -->|Yes| Q[Initiate Payment Processing]

    P --> B
    Q --> R[Send to Payment Gateway]
    R --> S{Payment Gateway Response}

    S -->|Processing| T[Show Processing Status<br/>- Real-time updates<br/>- Expected completion time<br/>- Transaction reference]
    S -->|Failed| U[Payment Failed<br/>- Insufficient funds<br/>- Network error<br/>- Gateway timeout<br/>- Security check failed]
    S -->|Success| V[Payment Successful]

    T --> W{Final Status Update}
    W -->|Failed| U
    W -->|Success| V

    U --> X{Retry Payment?}
    X -->|Yes| Y{Same Method?}
    Y -->|Yes| Q
    Y -->|No| B
    X -->|No| Z[Cancel Order]

    V --> AA[Update Order Status to PAID]
    AA --> BB[Send Payment Confirmation<br/>- Receipt generation<br/>- Transaction ID<br/>- Payment timestamp]

    BB --> CC[Notify Seller of Payment]
    CC --> DD[Trigger Fulfillment Process]
    DD --> EE[Add to Payment History]
    EE --> FF[End Payment Process]

    Z --> GG[Update Order Status to CANCELLED]
    GG --> FF

    style A fill:#e1f5fe
    style V fill:#c8e6c9
    style AA fill:#c8e6c9
    style BB fill:#c8e6c9
    style M fill:#ffcdd2
    style U fill:#ffcdd2
    style Z fill:#ffcdd2
```

## Cross-Workflow Integration Diagram

```mermaid
flowchart TD
    A[Product Discovery] --> B[Auction Participation]
    B --> C[Auto-Bidding Management]
    C --> D[Bid History Tracking]
    B --> D
    D --> E{Won Auction?}
    E -->|Yes| F[Order Creation]
    E -->|No| G[Continue Bidding]
    F --> H[Payment Processing]
    H --> I[Order Fulfillment]
    G --> A

    J[Real-time Notifications] --> B
    J --> C
    J --> F
    J --> H

    K[User Profile Management] --> A
    K --> B
    K --> F

    L[Security & Authentication] --> A
    L --> B
    L --> F
    L --> H

    style A fill:#e3f2fd
    style B fill:#f3e5f5
    style C fill:#e8f5e8
    style D fill:#fff3e0
    style F fill:#fce4ec
    style H fill:#f1f8e9
```
