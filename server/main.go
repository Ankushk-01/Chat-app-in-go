package main
import (
    "bufio"
    pb "chat/Chat/Chat"
    "context"
    "encoding/json"
    "errors"
    "flag"
    "fmt"
    "log"
    "net"
    "os"
    "strings"
    "sync"
    "time"
    "github.com/aws/aws-sdk-go/aws"
    // "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/lambda"
    "github.com/google/uuid"
    "google.golang.org/grpc"
)
type clientActivity struct {
    lastActivity time.Time
}
type Input struct {
    MatchKey string `json:"matchkey"`
    To string `json:"to"`
    BotType string `json:"botType"`
    Content string `json:"content"`
  }
type Server struct {
    rw sync.Mutex
    pb.UnimplementedMessageServer
    clients  map[string]clientInfo
    messages map[string]clientMessage // Store messages for each client
    channel  map[string]chan *pb.RecieveMessageResponse
    activity map[string]clientActivity
}
type clientInfo struct {
    name  string
    email string
}
type clientMessage struct {
    msgs   []string
    sender string
}
func (s *Server) updateActivity(clientID string) {
    if s.activity == nil {
        return
    }
    s.rw.Lock()
    fmt.Println("The Client id for activity is : ", clientID)
    s.activity[clientID] = clientActivity{lastActivity: time.Now()}
    s.rw.Unlock()
}
func checkActivity(s *Server, session int) {
    for {
        s.rw.Lock()
        if s.activity == nil {
            continue
        }
        for ID, clientActivity := range s.activity {
            if time.Since(clientActivity.lastActivity) >= time.Duration(session)*time.Minute {
                s.deleteClient(ID)
            }
        }
        s.rw.Unlock()
        time.Sleep(2 * time.Second)
    }
}
func newServer() *Server {
    s := Server{
        clients:  make(map[string]clientInfo),
        messages: make(map[string]clientMessage),
        channel:  make(map[string]chan *pb.RecieveMessageResponse),
        activity: make(map[string]clientActivity),
    }
    return &s
}
var (
    session2 = flag.Int("session", 10, "Session Timer")
    port     = flag.Int("port", 7000, "Server port")
)
func main() {
    flag.Parse()
    // Convert the session options object to an AWS config object.
    fmt.Println("Server App")
    // port := strconv
    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
    if err != nil {
        log.Fatalf("Unable to use port %v", err)
    }
    var opts []grpc.ServerOption
    grpcServer := grpc.NewServer(opts...)
    server := newServer()
    pb.RegisterMessageServer(grpcServer, server)
    go checkActivity(server, *session2)
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Unable to serve %v", err)
    }
}
func (s *Server) deleteClient(ClientId string) {
    for clientid := range s.clients {
        if strings.TrimSpace(clientid) == strings.TrimSpace(ClientId) {
            delete(s.clients, clientid)
            clientChannel := s.channel[clientid]
            deleteMessage := pb.RecieveMessageResponse{
                Sender: "Server",
                Text:   "Server is closing that connection because session is expired",
            }
            clientChannel <- &deleteMessage
            close(clientChannel)
            fmt.Println("Client is removed successfully")
        }
    }
}
func (s *Server) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.RegisterResponse, error) {
    // fmt.Println("Register method is started")
    // fmt.Println("Register method is called")
    clientName := request.GetUsername()
    clientEmail := request.GetEmail()
    clientInfo1 := clientInfo{
        name:  strings.TrimSpace(clientName),
        email: strings.TrimSpace(clientEmail),
    }
    // Check if the client already exists and return the existing ID if found
    for cID, cInfo := range s.clients {
        if clientInfo1.name == cInfo.name && clientInfo1.email == cInfo.email {
            return &pb.RegisterResponse{
                UserId: cID,
            }, nil
        }
    }
    // Generate a new client ID and add the client
    scanner := bufio.NewScanner(os.Stdin)
    fmt.Printf("\n Do you want to add a client whose name is %s? (yes or no): ", clientName)
    scanner.Scan()
    condition := scanner.Text()
    if strings.TrimSpace(strings.ToLower(condition)) == "yes" {
        fmt.Println("Client is allowed to join the chat")
        clientID := uuid.New().String()
        s.clients[clientID] = clientInfo1
        msgChennel := make(chan *pb.RecieveMessageResponse)
        s.channel[clientID] = msgChennel
        activity := clientActivity{lastActivity: time.Now()}
        s.activity[clientID] = activity
        fmt.Println("Active time ", activity.lastActivity)
        // fmt.Println("Client id is:", clientID)
        // fmt.Println("Client info:", clientInfo1)
        return &pb.RegisterResponse{
            UserId: clientID,
        }, nil
    }
    return &pb.RegisterResponse{
        UserId: uuid.Nil.String(),
    }, fmt.Errorf("Server not allowed to add client")
}
func (s *Server) SendMessage(stream pb.Message_SendMessageServer) error {
    sess := session.Must(session.NewSessionWithOptions(session.Options{
        SharedConfigState: session.SharedConfigEnable,
    }))
    client := lambda.New(sess, &aws.Config{Region: aws.String("ap-south-1")})
    // fmt.Println("Send message called")
    message, err := stream.Recv()
    if err != nil {
        stream.SendAndClose(&pb.SendMessageResponse{
            Status: false,
        })
        return errors.New(err.Error())
    }
    senderName := message.GetSender()
    // receiverName := message.GetReciever()
    msg := message.GetText()
    input := Input{
        MatchKey: msg,
        To:       senderName,
        BotType:  "FILE",
        Content:  "f",
    }
    payload, ok := json.Marshal(input);
    if ok != nil {
        fmt.Println("Error occurs when parsing the input : ")
    }
    result, err := client.Invoke(&lambda.InvokeInput{FunctionName: aws.String("decisionfunction-dev-decision"), Payload: payload})
    if err != nil {
        fmt.Println("Error calling decisionValue")
        os.Exit(0)
    }
    err = json.Unmarshal(result.Payload, &result)
    if err != nil {
        fmt.Println("Error unmarshalling MyGetItemsFunction response")
        os.Exit(0)
    }
    senderID := ""
    for cID, cInfo := range s.clients {
        if strings.TrimSpace(senderName) == cInfo.name {
            // fmt.Println("Cid of reciever: ",cID)
            senderID = cID
            break
        }
    }
    fmt.Printf("Sender ID : %v ", senderID)
    msgChannel := s.channel[senderID]
    msgChannel <- &pb.RecieveMessageResponse{
        Sender: senderName,
        Text:   result.GoString(),
    }
    s.updateActivity(senderID)
    // fmt.Println("Message is sended to the channel : ",msg)
    stream.SendAndClose(&pb.SendMessageResponse{
        Status: true,
    })
    return nil
}
func (s *Server) RecieveMessage(req *pb.RecieveMessageRequest, srv pb.Message_RecieveMessageServer) error {
    // fmt.Println("Recieve message called")
    clientID := req.GetClientID()
    if _, ok := s.clients[clientID]; !ok {
        return errors.New("CLient not found")
    }
    // msg := make(chan *pb.RecieveMessageResponse)
    msg := s.channel[clientID]
    for {
        select {
        case <-srv.Context().Done():
            {
                fMsg := &pb.RecieveMessageResponse{Text: "Stream is ended"}
                srv.Send(fMsg)
                return fmt.Errorf("Stream is Ended from client side")
            }
        case m1 := <-msg:
            {
                // fmt.Printf("GO ROUTINE (got message): %v \n", msg)
                srv.Send(m1)
            }
        }
    }
}