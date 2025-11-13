# Collaborator-Seller (Lister) Mermaid Workflow Diagrams

## Workflow 1: Product Listing Creation

```mermaid
flowchart TD
    A[Start Listing Creation] --> B[Select Product from Catalog]
    B --> C{Product Available?}
    C -->|No| D[Product Not Found Error]
    C -->|Yes| E[Check Inventory Availability]

    E --> F{Sufficient Inventory?}
    F -->|No| G[Insufficient Inventory Error]
    F -->|Yes| H[Configure Listing Details]

    H --> I[Set Quantity & Asking Price]
    I --> J[Define Minimum Bid]
    J --> K[Set Auction Duration]
    K --> L[Configure Pickup/Delivery Location]
    L --> M[Set Visibility Level]

    M --> N{Visibility Type}
    N -->|PUBLIC| O[Public Listing Configuration]
    N -->|PRIVATE| P[Private Listing Configuration]
    N -->|NETWORK| Q[Network Listing Configuration]
    N -->|ORGANIZATION| R[Organization Listing Configuration]

    P --> S[Configure Participant Invitations]
    S --> T[Set Bid Visibility Settings]
    O --> T
    Q --> T
    R --> T

    T --> U{Bid Visibility Level}
    U -->|FULL| V[All Bid Details Visible]
    U -->|PARTIAL| W[Limited Bid Information]
    U -->|MINIMAL| X[Basic Bid Count Only]
    U -->|HIDDEN| Y[No Bid Information]

    V --> Z[Add Terms & Conditions]
    W --> Z
    X --> Z
    Y --> Z

    Z --> AA[Review Listing Summary]
    AA --> BB{Confirm Listing?}
    BB -->|No| CC[Modify Listing Details]
    BB -->|Yes| DD[Submit Listing]

    CC --> H
    DD --> EE{Validation Successful?}
    EE -->|No| FF[Validation Errors<br/>- Invalid pricing<br/>- Missing required fields<br/>- Business rule violations]
    EE -->|Yes| GG[Create Active Listing]

    FF --> H
    GG --> HH[Generate Listing ID]
    HH --> II[Set Expiry Time]
    II --> JJ[Reserve Inventory]
    JJ --> KK[Send Listing Confirmation]
    KK --> LL[Notify Potential Buyers]
    LL --> MM[Begin Auction Monitoring]

    D --> B
    G --> B

    style A fill:#e1f5fe
    style GG fill:#c8e6c9
    style KK fill:#c8e6c9
    style D fill:#ffcdd2
    style G fill:#ffcdd2
    style FF fill:#ffcdd2
```

## Workflow 2: Auction Configuration & Management

```mermaid
flowchart TD
    A[Active Auction Created] --> B[Initialize Monitoring Dashboard]
    B --> C[Configure Real-time Updates]
    C --> D[Set Notification Preferences]

    D --> E[Monitor Auction Activity]
    E --> F{New Activity Detected?}
    F -->|No| G[Continue Monitoring]
    F -->|Yes| H{Activity Type}

    H -->|New Bid| I[Process Bid Notification]
    H -->|Bid Withdrawn| J[Handle Bid Withdrawal]
    H -->|Question from Bidder| K[Handle Bidder Inquiry]
    H -->|System Alert| L[Process System Alert]

    I --> M[Update Auction Dashboard]
    J --> M
    K --> N[Respond to Bidder]
    L --> O[Handle Alert Action]

    N --> M
    O --> M
    M --> P{Seller Action Required?}

    P -->|No| G
    P -->|Yes| Q{Action Type}

    Q -->|Extend Auction| R[Extend Auction Duration]
    Q -->|Modify Settings| S[Update Auction Parameters]
    Q -->|Close Early| T[Initiate Early Closure]
    Q -->|Add Participants| U[Send Additional Invitations]

    R --> V{Extension Valid?}
    V -->|No| W[Extension Error]
    V -->|Yes| X[Update Expiry Time]

    S --> Y{Settings Valid?}
    Y -->|No| Z[Settings Error]
    Y -->|Yes| AA[Apply New Settings]

    T --> BB{Early Closure Valid?}
    BB -->|No| CC[Closure Error]
    BB -->|Yes| DD[Close Auction]

    U --> EE{Invitations Valid?}
    EE -->|No| FF[Invitation Error]
    EE -->|Yes| GG[Send Invitations]

    W --> Q
    Z --> Q
    CC --> Q
    FF --> Q

    X --> HH[Notify Participants of Extension]
    AA --> II[Notify Participants of Changes]
    DD --> JJ[Process Auction Closure]
    GG --> KK[Track Invitation Responses]

    HH --> G
    II --> G
    JJ --> LL[End Auction Management]
    KK --> G
    G --> E

    style A fill:#e1f5fe
    style B fill:#e8f5e8
    style DD fill:#c8e6c9
    style W fill:#ffcdd2
    style Z fill:#ffcdd2
    style CC fill:#ffcdd2
    style FF fill:#ffcdd2
```

## Workflow 3: Listing Visibility Control

```mermaid
flowchart TD
    A[Configure Listing Visibility] --> B{Visibility Level Selection}

    B -->|PUBLIC| C[Public Listing Setup]
    B -->|PRIVATE| D[Private Listing Setup]
    B -->|NETWORK| E[Network Listing Setup]
    B -->|ORGANIZATION| F[Organization Listing Setup]

    C --> G[Allow All Authenticated Users]
    G --> H[Configure Bid Transparency]

    D --> I[Create Invitation List]
    I --> J[Send Participant Invitations]
    J --> K[Track Invitation Responses]
    K --> L[Manage Participant Access]
    L --> H

    E --> M[Identify Partner Organizations]
    M --> N[Validate Network Memberships]
    N --> O[Enable Network Access]
    O --> H

    F --> P[Restrict to Organization Members]
    P --> Q[Validate Organization Membership]
    Q --> R[Enable Organization Access]
    R --> H

    H --> S{Bid Transparency Level}
    S -->|FULL| T[Show All Bid Details<br/>- Bid amounts<br/>- Timestamps<br/>- Anonymous bidder IDs]
    S -->|PARTIAL| U[Show Limited Information<br/>- Current highest bid<br/>- Total bid count]
    S -->|MINIMAL| V[Show Basic Information<br/>- Bid count only]
    S -->|HIDDEN| W[Hide All Bid Information<br/>- No bid details during auction]

    T --> X[Apply Visibility Settings]
    U --> X
    V --> X
    W --> X

    X --> Y[Validate Access Controls]
    Y --> Z{Validation Successful?}
    Z -->|No| AA[Access Control Error<br/>- Invalid participant list<br/>- Network permission issues<br/>- Organization access problems]
    Z -->|Yes| BB[Activate Visibility Rules]

    AA --> B
    BB --> CC[Monitor Access Compliance]
    CC --> DD{Access Violation Detected?}
    DD -->|Yes| EE[Handle Access Violation<br/>- Log unauthorized access<br/>- Block violating user<br/>- Notify seller]
    DD -->|No| FF[Continue Monitoring]

    EE --> GG[Update Security Settings]
    GG --> FF
    FF --> CC

    J --> HH{Invitation Response}
    HH -->|Accepted| II[Grant Participant Access]
    HH -->|Declined| JJ[Remove from Participant List]
    HH -->|No Response| KK[Send Reminder]

    II --> LL[Update Participant Database]
    JJ --> LL
    KK --> MM[Track Reminder Status]
    MM --> LL
    LL --> K

    style A fill:#e1f5fe
    style BB fill:#c8e6c9
    style LL fill:#c8e6c9
    style AA fill:#ffcdd2
    style EE fill:#fff3e0
```

## Workflow 4: Bid Monitoring & Analysis

```mermaid
flowchart TD
    A[Auction Active] --> B[Initialize Bid Monitoring]
    B --> C[Set Up Real-time Dashboard]
    C --> D[Configure Analytics Tracking]

    D --> E[Monitor Incoming Bids]
    E --> F{New Bid Received?}
    F -->|No| G[Check Dashboard Updates]
    F -->|Yes| H[Process Bid Information]

    H --> I{Bid Visibility Settings}
    I -->|FULL| J[Display Complete Bid Details<br/>- Bid amount<br/>- Bidder identifier<br/>- Timestamp<br/>- Quantity requested]
    I -->|PARTIAL| K[Display Limited Information<br/>- New highest bid indicator<br/>- Updated bid count]
    I -->|MINIMAL| L[Display Basic Update<br/>- Bid count increment]
    I -->|HIDDEN| M[No Bid Display<br/>- Internal tracking only]

    J --> N[Update Bid Analytics]
    K --> N
    L --> N
    M --> N

    N --> O[Analyze Bid Patterns<br/>- Bid progression trends<br/>- Bidder engagement<br/>- Price discovery patterns<br/>- Competition intensity]

    O --> P[Update Performance Metrics]
    P --> Q{Generate Insights?}
    Q -->|Yes| R[Generate Market Intelligence<br/>- Demand assessment<br/>- Price optimization suggestions<br/>- Bidder behavior analysis<br/>- Competition benchmarks]
    Q -->|No| S[Continue Monitoring]

    R --> T[Display Insights Dashboard]
    T --> U{Seller Action Required?}
    U -->|Yes| V{Action Type}
    U -->|No| S

    V -->|Respond to Bidder| W[Send Bidder Response]
    V -->|Adjust Strategy| X[Modify Auction Parameters]
    V -->|Extend Duration| Y[Extend Auction Time]
    V -->|Add Information| Z[Update Product Details]

    W --> AA[Track Communication]
    X --> BB[Validate Strategy Changes]
    Y --> CC[Process Duration Extension]
    Z --> DD[Update Listing Information]

    AA --> EE{Communication Successful?}
    BB --> FF{Changes Valid?}
    CC --> GG{Extension Approved?}
    DD --> HH{Update Successful?}

    EE -->|No| II[Communication Error]
    EE -->|Yes| JJ[Log Communication]
    FF -->|No| KK[Strategy Change Error]
    FF -->|Yes| LL[Apply Strategy Changes]
    GG -->|No| MM[Extension Error]
    GG -->|Yes| NN[Apply Extension]
    HH -->|No| OO[Update Error]
    HH -->|Yes| PP[Confirm Update]

    II --> V
    KK --> V
    MM --> V
    OO --> V

    JJ --> S
    LL --> S
    NN --> S
    PP --> S
    S --> G

    G --> QQ{Auction Status Check}
    QQ -->|Active| E
    QQ -->|Expired| RR[Process Auction Completion]
    QQ -->|Closed| RR

    RR --> SS[Generate Final Analytics Report]
    SS --> TT[End Monitoring]

    style A fill:#e1f5fe
    style C fill:#e8f5e8
    style R fill:#c8e6c9
    style SS fill:#c8e6c9
    style II fill:#ffcdd2
    style KK fill:#ffcdd2
    style MM fill:#ffcdd2
    style OO fill:#ffcdd2
```

## Workflow 5: Auction Results Management

```mermaid
flowchart TD
    A[Auction Expiry Time Reached] --> B{Auction Closure Type}
    B -->|Automatic| C[System Auto-Close]
    B -->|Manual| D[Seller Initiated Close]
    B -->|Admin Force Close| E[Admin Intervention]

    C --> F[Validate Auction State]
    D --> G[Validate Close Permission]
    E --> H[Process Admin Close]

    F --> I{Validation Result}
    G --> I
    H --> I

    I -->|Failed| J[Close Validation Error<br/>- Active bids in process<br/>- System conflicts<br/>- Permission issues]
    I -->|Passed| K[Determine Winning Bid]

    J --> L[Handle Closure Error]
    L --> M{Retry Closure?}
    M -->|Yes| F
    M -->|No| N[Escalate to Admin]

    K --> O{Bids Exist?}
    O -->|No| P[No Bids Scenario]
    O -->|Yes| Q[Identify Highest Valid Bid]

    P --> R[Mark as EXPIRED_NO_BIDS]
    R --> S[Notify Seller of No Bids]
    S --> T[Release Reserved Inventory]
    T --> U[Generate No-Bid Report]

    Q --> V{Multiple Same Amount Bids?}
    V -->|Yes| W[Apply Tie-Breaking Rules<br/>- First bid wins (timestamp)<br/>- Higher quantity preference<br/>- Bidder rating consideration]
    V -->|No| X[Declare Winner]

    W --> X
    X --> Y[Validate Winning Bid]
    Y --> Z{Winner Validation}
    Z -->|Failed| AA[Winner Validation Error<br/>- Bid no longer valid<br/>- Bidder account issues<br/>- Payment verification failed]
    Z -->|Passed| BB[Confirm Auction Winner]

    AA --> CC[Select Next Valid Bid]
    CC --> Y

    BB --> DD[Update Auction Status to CLOSED]
    DD --> EE[Notify Winner]
    EE --> FF[Notify Other Bidders]
    FF --> GG[Notify Seller]

    GG --> HH[Generate Auction Results Report<br/>- Final winning bid<br/>- Total bid count<br/>- Bid progression<br/>- Performance metrics]

    HH --> II[Create Order Opportunity]
    II --> JJ[Set Order Creation Deadline]
    JJ --> KK[Monitor Order Creation]

    KK --> LL{Order Created by Winner?}
    LL -->|Yes| MM[Process Order]
    LL -->|No| NN{Deadline Expired?}
    NN -->|No| KK
    NN -->|Yes| OO[Offer to Next Bidder]

    MM --> PP[Begin Fulfillment Process]
    OO --> QQ{Next Bidder Accepts?}
    QQ -->|Yes| MM
    QQ -->|No| RR[Continue to Next Bidder]
    RR --> OO

    PP --> SS[Update Seller Performance Metrics]
    U --> SS
    SS --> TT[Archive Auction Data]
    TT --> UU[End Results Management]

    N --> VV[Admin Intervention Required]
    VV --> UU

    style A fill:#e1f5fe
    style BB fill:#c8e6c9
    style HH fill:#c8e6c9
    style PP fill:#c8e6c9
    style J fill:#ffcdd2
    style AA fill:#ffcdd2
    style N fill:#fff3e0
    style VV fill:#fff3e0
```

## Workflow 6: Order Fulfillment

```mermaid
flowchart TD
    A[Order Created from Winning Bid] --> B[Receive Order Notification]
    B --> C[Validate Order Details<br/>- Product specifications<br/>- Quantity availability<br/>- Pricing accuracy<br/>- Buyer information]

    C --> D{Order Validation}
    D -->|Failed| E[Order Validation Error<br/>- Inventory shortfall<br/>- Pricing discrepancy<br/>- Buyer verification failed]
    D -->|Passed| F[Review Order Terms]

    E --> G{Resolve Issues?}
    G -->|Yes| H[Correct Order Issues]
    G -->|No| I[Reject Order]

    H --> C
    I --> J[Notify Buyer of Rejection]
    J --> K[Offer to Next Bidder]

    F --> L{Accept Order?}
    L -->|No| M[Decline Order]
    L -->|Yes| N[Accept Order]

    M --> O[Provide Decline Reason]
    O --> J

    N --> P[Update Order Status to CONFIRMED]
    P --> Q[Allocate Specific Inventory]
    Q --> R[Prepare Fulfillment Documentation]

    R --> S[Quality Check & Preparation<br/>- Product quality verification<br/>- Packaging preparation<br/>- Documentation completion<br/>- Certification validation]

    S --> T{Quality Check Passed?}
    T -->|No| U[Quality Issues Identified]
    T -->|Yes| V[Arrange Logistics]

    U --> W{Resolve Quality Issues?}
    W -->|Yes| X[Address Quality Problems]
    W -->|No| Y[Cancel Order Due to Quality]

    X --> S
    Y --> Z[Notify Buyer of Cancellation]
    Z --> K

    V --> AA{Delivery Method}
    AA -->|Pickup| BB[Coordinate Pickup Schedule]
    AA -->|Delivery| CC[Arrange Delivery Service]

    BB --> DD[Notify Buyer of Pickup Details<br/>- Location and address<br/>- Available time slots<br/>- Contact information<br/>- Required documentation]

    CC --> EE[Schedule Delivery<br/>- Delivery address confirmation<br/>- Delivery time window<br/>- Logistics partner coordination<br/>- Tracking setup]

    DD --> FF[Prepare for Pickup]
    EE --> GG[Prepare for Shipment]

    FF --> HH[Buyer Pickup Process<br/>- Identity verification<br/>- Product inspection<br/>- Documentation exchange<br/>- Payment confirmation]

    GG --> II[Shipment Dispatch<br/>- Package handover to carrier<br/>- Tracking information generation<br/>- Delivery confirmation setup<br/>- Insurance and protection]

    HH --> JJ{Pickup Successful?}
    II --> KK{Shipment Dispatched?}

    JJ -->|No| LL[Handle Pickup Issues<br/>- Reschedule pickup<br/>- Resolve documentation issues<br/>- Address quality concerns]
    JJ -->|Yes| MM[Confirm Pickup Completion]

    KK -->|No| NN[Handle Shipment Issues<br/>- Resolve logistics problems<br/>- Coordinate with carrier<br/>- Update delivery schedule]
    KK -->|Yes| OO[Track Delivery Progress]

    LL --> FF
    NN --> GG

    MM --> PP[Process Payment Completion]
    OO --> QQ{Delivery Completed?}
    QQ -->|No| RR[Monitor Delivery Status]
    QQ -->|Yes| SS[Confirm Delivery Completion]

    RR --> OO
    SS --> PP

    PP --> TT[Update Order Status to COMPLETED]
    TT --> UU[Generate Completion Documentation<br/>- Delivery receipt<br/>- Payment confirmation<br/>- Quality certification<br/>- Performance metrics]

    UU --> VV[Request Buyer Feedback]
    VV --> WW[Update Seller Performance Metrics]
    WW --> XX[Archive Order Documentation]
    XX --> YY[Release Remaining Inventory]
    YY --> ZZ[End Fulfillment Process]

    K --> AAA[End Alternative Path]
    AAA --> ZZ

    style A fill:#e1f5fe
    style N fill:#c8e6c9
    style MM fill:#c8e6c9
    style SS fill:#c8e6c9
    style UU fill:#c8e6c9
    style E fill:#ffcdd2
    style U fill:#fff3e0
    style Y fill:#ffcdd2
    style LL fill:#fff3e0
    style NN fill:#fff3e0
```

## Cross-Workflow Integration Diagram

```mermaid
flowchart TD
    A[Product Listing Creation] --> B[Auction Configuration & Management]
    B --> C[Listing Visibility Control]
    C --> D[Bid Monitoring & Analysis]
    D --> E[Auction Results Management]
    E --> F[Order Fulfillment]

    G[Inventory Management] --> A
    G --> E
    G --> F

    H[Communication System] --> B
    H --> C
    H --> D
    H --> F

    I[Analytics & Reporting] --> D
    I --> E
    I --> F

    J[Payment Processing] --> F
    K[Quality Assurance] --> F
    L[Logistics Management] --> F

    M[User Management] --> A
    M --> B
    M --> C

    N[Security & Authentication] --> A
    N --> B
    N --> C
    N --> D

    style A fill:#e3f2fd
    style B fill:#f3e5f5
    style C fill:#e8f5e8
    style D fill:#fff3e0
    style E fill:#fce4ec
    style F fill:#f1f8e9
```
