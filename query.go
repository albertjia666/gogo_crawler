package main

import (
        "encoding/json"
        "fmt"
        "net/http"
        "net/http/cookiejar"
        "net/url"
        "os"
        "strconv"
        "strings"
        "time"

        "github.com/robfig/cron"

        jira "github.com/andygrunwald/go-jira"
        "github.com/antchfx/htmlquery"
        "github.com/cyberark/conjur-api-go/conjurapi"
        "github.com/cyberark/conjur-api-go/conjurapi/authn"
        "github.com/nlopes/slack"
        "github.com/sirupsen/logrus"
        "github.com/trivago/tgo/tcontainer"
        "golang.org/x/net/html"
)

func goQeury(urlStr string) *html.Node {
        fmt.Println("=============================")
        fmt.Println("Time now:", time.Now().Format("2006-01-02 15:04:05"))
        fmt.Println("=============================")
        jar, _ := cookiejar.New(nil)
        client := &http.Client{
                Jar:     jar,
                Timeout: time.Second * 60,
        }
        // Get Target URL Cookies
        req1, req1Err := http.NewRequest("GET", "https://xxxxxxxxxxxxxxxxxxxxxxxxxxxxx", nil)
        if req1Err != nil {
                fmt.Printf("req1 fetch error: %v\n", req1Err)
                return nil
        }
        res1, res1Err := client.Do(req1)
        if res1Err != nil {
                fmt.Printf("res1 fetch error: %v\n", res1Err)
                return nil
        }
        defer res1.Body.Close()
        doc1, doc1Err := htmlquery.Parse(res1.Body)
        if doc1Err != nil {
                fmt.Printf("doc1 parse error: %v\n", doc1Err)
                return nil
        }
        token := htmlquery.FindOne(doc1, "//input[@name='_token']")
        tokenStr := token.Attr[2].Val

        println("Step1 GET Request Complete...")

        // Login Target URL with Post Data & Cookies
        postValues := url.Values{}
        postValues.Add("_token", tokenStr)
        postValues.Add("email", os.Getenv("MAIL"))
        postValues.Add("password", os.Getenv("PASSWORD"))
        postValues.Add("remember", "on")
        req2, req2Err := http.NewRequest("POST", "https://xxxxxxxxxxxxxxxxxxxxxxxxxxxxx", strings.NewReader(postValues.Encode()))
        if req2Err != nil {
                fmt.Printf("req2 fetch error: %v\n", req2Err)
                return nil
        }
        req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
        req2.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36")
        res2, res2Err := client.Do(req2)
        if res2Err != nil {
                fmt.Printf("res2 fetch error: %v\n", res2Err)
                return nil
        }
        defer res2.Body.Close()

        println("Step2 POST Request Complete...")

        req3, req3Err := http.NewRequest("GET", urlStr, nil)
        if req3Err != nil {
                fmt.Printf("req3 fetch error: %v\n", req3Err)
                return nil
        }

        res3, res3Err := client.Do(req3)
        if res3Err != nil {
                fmt.Printf("res3 fetch error: %v\n", res3Err)
                return nil
        }
        defer res3.Body.Close()

        println("Step3 GET Request For Info Complete...")

        //body3, _ := ioutil.ReadAll(res3.Body)
        //fmt.Println(string(body3))
        doc3, _ := htmlquery.Parse(res3.Body)
        return doc3
}

func goQueryAxc() {
        urlStr := "https://xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
        doc3 := goQeury(urlStr)
        checkTitle := htmlquery.FindOne(doc3, "//h1[@class='display-4']//text()").Data
        fmt.Printf("%#v\n", checkTitle)
        // checkTime := htmlquery.FindOne(doc3, "//span[@class='glyphicon glyphicon-refresh']//text()").Data
        // fmt.Printf("%#v\n", strings.TrimSpace(checkTime))
        checkTime := time.Now().Format("2006-01-02 15:04:05")
        fmt.Printf("%#v\n", strings.TrimSpace(checkTime))

        cardMap := make(map[string][]string)
        cardItem := htmlquery.Find(doc3, "//div[contains(@class,'mx-auto')]")
        fmt.Printf("%#v\n", len(cardItem))
        for index, value := range cardItem {
                fmt.Printf("%#v\n", index)
                fmt.Printf("%#v\n", value)
                cardItemTitle := htmlquery.FindOne(value, "//div[@class='card-header text-center']//text()").Data
                fmt.Printf("%#v\n", cardItemTitle)
                carList := make([]string, 0)
                cardItemInfos := htmlquery.Find(value, "//div[@class='col']")
                for _, value := range cardItemInfos {
                        //fmt.Printf("%#v\n", index)
                        infoText := htmlquery.FindOne(value, "//div[@class='col']//text()").Data
                        //fmt.Printf("%#v\n", strings.TrimSpace(infoText))
                        carList = append(carList, strings.TrimSpace(infoText))
                }
                cardMap[cardItemTitle] = carList
        }
        fmt.Printf("%#v\n", cardMap)
        mapJSONStr, _ := json.Marshal(cardMap)
        fmt.Printf("%s\n", string(mapJSONStr))

        allCampaign := htmlquery.Find(doc3, "//table[@id='ax-level']/tbody/tr")
        allCampaignCount := len(allCampaign)
        fmt.Printf("%#v\n", len(allCampaign))
        fmt.Printf("%#v\n", allCampaignCount)
        //fmt.Printf("%#v\n", allCampaign)
        campaignFailList := make([]map[string]string, 0)
        campaignNotCompleteList := make([]map[string]string, 0)
        campaignNotPostList := make([]map[string]string, 0)
        campaignProcessedList := make([]map[string]string, 0)
        for _, value := range allCampaign {
                //fmt.Printf("%#v\n", value)
                campaignMap := make(map[string]string)
                campaignItem := htmlquery.Find(value, "//td")
                campaignMap["gwyName"] = strings.TrimSpace(htmlquery.InnerText(campaignItem[1]))
                campaignMap["gwyIpAddr"] = strings.TrimSpace(htmlquery.InnerText(campaignItem[2]))
                campaignMap["campaignLSID"] = strings.TrimSpace(htmlquery.InnerText(campaignItem[3]))
                campaignMap["campaignGSID"] = strings.TrimSpace(htmlquery.InnerText(campaignItem[4]))
                campaignMap["campaignPackId"] = strings.TrimSpace(htmlquery.InnerText(campaignItem[5]))
                campaignMap["campaignCmdStatus"] = strings.TrimSpace(htmlquery.InnerText(campaignItem[9]))
                campaignMap["campaignGwyStatus"] = strings.TrimSpace(htmlquery.InnerText(campaignItem[10]))
                campaignMap["campaignTsSteps"] = strings.TrimSpace(htmlquery.InnerText(campaignItem[12]))
                fmt.Printf("%#v\n", campaignMap)
                if campaignMap["campaignCmdStatus"] == "FAILED" {
                        campaignFailList = append(campaignFailList, campaignMap)
                }
                if (campaignMap["campaignCmdStatus"] == "SUCCESS" || strings.Contains(campaignMap["campaignCmdStatus"], "%")) && campaignMap["campaignGwyStatus"] != "PROCESSED" {
                        campaignNotCompleteList = append(campaignNotCompleteList, campaignMap)
                }
                if campaignMap["campaignPackId"] == "XXXXXX" {
                        campaignNotPostList = append(campaignNotPostList, campaignMap)
                }
                if campaignMap["campaignCmdStatus"] == "SUCCESS" && campaignMap["campaignGwyStatus"] == "PROCESSED" {
                        campaignProcessedList = append(campaignProcessedList, campaignMap)
                }
        }
        fmt.Printf("===>>> %d %#v\n", len(campaignFailList), campaignFailList)
        fmt.Printf("===>>> %d %#v\n", len(campaignNotCompleteList), campaignNotCompleteList)
        fmt.Printf("===>>> %d %#v\n", len(campaignNotPostList), campaignNotPostList)
        fmt.Printf("===>>> %d %#v\n", len(campaignProcessedList), campaignProcessedList)

        AxcDataHandle(allCampaignCount, checkTitle, checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList)
}

func goQuerySoc() {
        urlStr := "https://xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
        doc3 := goQeury(urlStr)
        checkTitle := htmlquery.FindOne(doc3, "//h1[@class='display-4']//text()").Data
        fmt.Printf("%#v\n", checkTitle)
        checkTime := htmlquery.FindOne(doc3, "//span[@class='glyphicon glyphicon-refresh']//text()").Data
        fmt.Printf("%#v\n", strings.TrimSpace(checkTime))

        allCampaign := htmlquery.Find(doc3, "//table[@id='so-level']/tbody/tr")
        allCampaignCount := len(allCampaign)
        fmt.Printf("%#v\n", allCampaign)
        fmt.Printf("%#v\n", allCampaignCount)
        campaignFailList := make([]map[string]string, 0)
        campaignNotCompleteList := make([]map[string]string, 0)
        campaignNotPostList := make([]map[string]string, 0)
        campaignProcessedList := make([]map[string]string, 0)
        for _, value := range allCampaign {
                //fmt.Printf("%#v\n", value)
                campaignMap := make(map[string]string)
                campaignItem := htmlquery.Find(value, "//td")
                campaignMap["gwyName"] = strings.TrimSpace(htmlquery.InnerText(campaignItem[1]))
                campaignMap["gwyIpAddr"] = strings.TrimSpace(htmlquery.InnerText(campaignItem[2]))
                campaignMap["campaignLSID"] = strings.TrimSpace(htmlquery.InnerText(campaignItem[3]))
                campaignMap["campaignGSID"] = strings.TrimSpace(htmlquery.InnerText(campaignItem[4]))
                campaignMap["campaignPackId"] = strings.TrimSpace(htmlquery.InnerText(campaignItem[5]))
                campaignMap["campaignCmdStatus"] = strings.TrimSpace(htmlquery.InnerText(campaignItem[9]))
                campaignMap["campaignGwyStatus"] = strings.TrimSpace(htmlquery.InnerText(campaignItem[10]))
                campaignMap["campaignTsSteps"] = strings.TrimSpace(htmlquery.InnerText(campaignItem[12]))
                fmt.Printf("%#v\n", campaignMap)
                if campaignMap["campaignCmdStatus"] == "FAILED" {
                        campaignFailList = append(campaignFailList, campaignMap)
                }
                if (campaignMap["campaignCmdStatus"] == "SUCCESS" || strings.Contains(campaignMap["campaignCmdStatus"], "%")) && campaignMap["campaignGwyStatus"] != "PROCESSED" {
                        campaignNotCompleteList = append(campaignNotCompleteList, campaignMap)
                }
                if campaignMap["campaignCmdStatus"] == "NONE" {
                        campaignNotPostList = append(campaignNotPostList, campaignMap)
                }
                if campaignMap["campaignCmdStatus"] == "SUCCESS" && campaignMap["campaignGwyStatus"] == "PROCESSED" {
                        campaignProcessedList = append(campaignProcessedList, campaignMap)
                }
        }

        SocDataHandle(allCampaignCount, checkTitle, checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList)
}

func AxcDataHandle(allCampaignCount int, checkTitle, checkTime string, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList []map[string]string) {
        logrus.Info("[QUERY] AxcSlackHandle Start...")
        logrus.Info("[QUERY] checkTitle: ", checkTitle)
        logrus.Info("[QUERY] checkTime: ", checkTime)
        fmt.Printf("===>>> %d %#v\n", len(campaignFailList), campaignFailList)
        fmt.Printf("===>>> %d %#v\n", len(campaignNotCompleteList), campaignNotCompleteList)
        fmt.Printf("===>>> %d %#v\n", len(campaignNotPostList), campaignNotPostList)
        fmt.Printf("===>>> %d %#v\n", len(campaignProcessedList), campaignProcessedList)
        utcHour := time.Now().Hour()
        utcMinute := time.Now().Minute()
        fmt.Printf("%d\n", utcHour)
        fmt.Printf("%d\n", utcMinute)

        // AxcJiraHandle(checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList)

        if len(campaignNotPostList) == allCampaignCount {
                // ALL PKGs NOT POST & NEED JIRA TICKET
                if utcHour == 14 && utcMinute >= 1 {
                        ticket := AxcJiraHandle(checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList)
                        AxcSlackHandle(checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList, 11, ticket)
                        AxcChannelPost(checkTime, ticket)
                } else {
                        AxcSlackHandle(checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList, 1)
                }
        } else if len(campaignFailList) == 0 && len(campaignNotCompleteList) == 0 && len(campaignNotPostList) == 0 && len(campaignProcessedList) == allCampaignCount {
                // ALL PKGs POST SUCCESS
                AxcSlackHandle(checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList, 0)
        } else {
                // PKGs(FAIL & NOT COMPLETE & NOT POST) & NEED JIRA TICKET
                if utcHour == 14 && utcMinute >= 1 {
                        ticket := AxcJiraHandle(checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList)
                        AxcSlackHandle(checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList, 22, ticket)
                        AxcChannelPost(checkTime, ticket)
                } else {
                        AxcSlackHandle(checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList, 2)
                }
        }
}

// SocDataHandle for SO check data parse
func SocDataHandle(allCampaignCount int, checkTitle, checkTime string, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList []map[string]string) {
        logrus.Info("[QUERY] SocSlackHandle Start...")
        logrus.Info("[QUERY] checkTitle: ", checkTitle)
        logrus.Info("[QUERY] checkTime: ", checkTime)
        fmt.Printf("===>>> %d %#v\n", len(campaignFailList), campaignFailList)
        fmt.Printf("===>>> %d %#v\n", len(campaignNotCompleteList), campaignNotCompleteList)
        fmt.Printf("===>>> %d %#v\n", len(campaignNotPostList), campaignNotPostList)
        fmt.Printf("===>>> %d %#v\n", len(campaignProcessedList), campaignProcessedList)
        utcHour := time.Now().Hour()
        utcMinute := time.Now().Minute()
        fmt.Printf("%d\n", utcHour)
        fmt.Printf("%d\n", utcMinute)

        // SocJiraHandle(checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList)

        if len(campaignNotPostList) == allCampaignCount {
                if utcHour == 14 && utcMinute >= 24 {
                        ticket := SocJiraHandle(checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList)
                        SocSlackHandle(checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList, 11, ticket)
                        SocChannelPost(checkTime, ticket)
                } else {
                        SocSlackHandle(checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList, 1)
                }
        } else if len(campaignFailList) == 0 && len(campaignNotCompleteList) == 0 && len(campaignNotPostList) == 0 && len(campaignProcessedList) == allCampaignCount {
                // ALL PKGs POST SUCCESS
                SocSlackHandle(checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList, 0)
        } else {
                if utcHour == 14 && utcMinute >= 24 {
                        ticket := SocJiraHandle(checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList)
                        SocSlackHandle(checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList, 22, ticket)
                        SocChannelPost(checkTime, ticket)
                } else {
                        SocSlackHandle(checkTime, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList, 2)
                }
        }

}

// JiraHandle to get jira client
func jiraHandle() (client *jira.Client, err error) {
        logrus.Info("[QUERY] JiraHandle Start...")
        jiraUser, jiraPsw := getJiraAuth()
        tp := jira.BasicAuthTransport{
                Username: string(jiraUser),
                Password: string(jiraPsw),
        }
        client, err = jira.NewClient(tp.Client(), "https://xxxxxxxxxxxxxxxxxxxxxxx")
        if err != nil {
                logrus.Error("Get Jira Handle Error: ", err)
        }
        //issue, _, _ := client.Issue.Get("ESC-xxxxx", nil)
        //fmt.Printf("%s: %+v\n", issue.Key, issue.Fields.Summary)
        //fmt.Printf("Type: %s\n", issue.Fields.Type.Name)
        //fmt.Printf("Priority: %s\n", issue.Fields.Priority.Name)
        logrus.Info("[QUERY] JiraHandle Client: ", &client)
        return client, err
}

// jiraDescriptionHandle for jira description
func jiraDescriptionHandle(campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList []map[string]string) (jiraDescription string) {
        logrus.Info("[QUERY] jiraDescriptionHandle Start...")
        jiraDesTblHeader := "||Gateway||Gateway IP||LSID||GSID||PackID||\r\n"
        var (
                jiraDesTblFail        string
                jiraDesTblNotComplete string
                jiraDesTblNotPost     string
        )
        if len(campaignFailList) > 0 {
                jiraDesTblFailSub := "h3.Packages Fail\r\n"
                jiraDesTblFailRows := ""
                for _, item := range campaignFailList {
                        jiraDesTblFailRow := fmt.Sprintf("|%s|%s|%s|%s|%s|\r\n", item["gwyName"], item["gwyIpAddr"], item["campaignLSID"], item["campaignGSID"], item["campaignPackId"])
                        jiraDesTblFailRows = jiraDesTblFailRows + jiraDesTblFailRow
                }
                jiraDesTblFail = jiraDesTblFailSub + jiraDesTblHeader + jiraDesTblFailRows + "\r\n"
        }
        if len(campaignNotCompleteList) > 0 {
                jiraDesTblNotCompleteSub := "h3.Packages Not Complete\r\n"
                jiraDesTblNotCompleteRows := ""
                for _, item := range campaignNotCompleteList {
                        jiraDesTblNotCompleteRow := fmt.Sprintf("|%s|%s|%s|%s|%s|\r\n", item["gwyName"], item["gwyIpAddr"], item["campaignLSID"], item["campaignGSID"], item["campaignPackId"])
                        jiraDesTblNotCompleteRows = jiraDesTblNotCompleteRows + jiraDesTblNotCompleteRow
                }
                jiraDesTblNotComplete = jiraDesTblNotCompleteSub + jiraDesTblHeader + jiraDesTblNotCompleteRows + "\r\n"
        }
        if len(campaignNotPostList) > 0 {
                jiraDesTblNotPostSub := "h3.Packages Not Post\r\n"
                jiraDesTblNotPostRows := ""
                for _, item := range campaignNotPostList {
                        jiraDesTblNotPostRow := fmt.Sprintf("|%s|%s|%s|%s|%s|\r\n", item["gwyName"], item["gwyIpAddr"], item["campaignLSID"], item["campaignGSID"], item["campaignPackId"])
                        jiraDesTblNotPostRows = jiraDesTblNotPostRows + jiraDesTblNotPostRow
                }
                jiraDesTblNotPost = jiraDesTblNotPostSub + jiraDesTblHeader + jiraDesTblNotPostRows + "\r\n"
        }
        jiraDescription = jiraDesTblFail + jiraDesTblNotComplete + jiraDesTblNotPost
        return jiraDescription
}

// AxcJiraHandle for AX jira handle
func AxcJiraHandle(checkTime string, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList []map[string]string) (ticket string) {
        logrus.Info("[QUERY] AxcJiraHandle Start...")
        client, err := jiraHandle()
        if err != nil {
                return
        }
        fmt.Printf("%#v\n", client)
        checkTime = strings.Replace(strings.Split(strings.TrimSpace(checkTime), " ")[0], "-", "/", -1)
        jiraSummary := fmt.Sprintf("Packages NOT completed for %s daily check", checkTime)
        fmt.Printf("%s\n", jiraSummary)
        jiraDescription := jiraDescriptionHandle(campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList)
        fmt.Println(jiraDescription)
        customfields := tcontainer.NewMarshalMap()
        customfields["customfield_14444"] = map[string]string{"value": "xxxxxx"}
        customfields["customfield_10212"] = []map[string]string{{"value": "xxxxxx"}}
        customfields["customfield_11919"] = []map[string]string{{"value": "AX"}, {"value": "Campaign"}, {"value": "Gateway"}}
        customfields["customfield_11920"] = []map[string]string{}
        fmt.Printf("%#v\n", customfields)
        i := jira.Issue{
                Fields: &jira.IssueFields{
                        Project: jira.Project{
                                Key: "ESC",
                        },
                        Type: jira.IssueType{
                                Name: "Investigation",
                        },
                        Priority: &jira.Priority{
                                Name: "NORMAL - P3",
                        },
                        Assignee: &jira.User{
                                Name: "xxxxxx",
                        },
                        Components: []*jira.Component{
                                {
                                        Name: "xxxxxx",
                                },
                        },
                        Description: jiraDescription,
                        Summary:     jiraSummary,
                        Unknowns:    customfields,
                },
        }
        issue, _, err := client.Issue.Create(&i)
        if err != nil {
                fmt.Println(err)
        }
        ticket = issue.Key
        fmt.Println(ticket)
        fmt.Println("Done")
        return ticket
}

// SocJiraHandle for SO jira handle
func SocJiraHandle(checkTime string, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList []map[string]string) (ticket string) {
        logrus.Info("[QUERY] SocJiraHandle Start...")
        client, err := jiraHandle()
        if err != nil {
                return
        }
        fmt.Printf("%#v\n", client)
        checkTime = strings.Replace(strings.Split(strings.TrimSpace(checkTime), " ")[0], "-", "/", -1)
        jiraSummary := fmt.Sprintf("Packages NOT completed for %s daily check", checkTime)
        fmt.Printf("%s\n", jiraSummary)
        jiraDescription := jiraDescriptionHandle(campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList)
        fmt.Println(jiraDescription)
        customfields := tcontainer.NewMarshalMap()
        customfields["customfield_14444"] = map[string]string{"value": "Internal Escalation"}
        customfields["customfield_10212"] = []map[string]string{{"value": "xxxxxx"}}
        customfields["customfield_11919"] = []map[string]string{{"value": "xxxxxx"}}
        customfields["customfield_11920"] = []map[string]string{}
        fmt.Printf("%#v\n", customfields)
        i := jira.Issue{
                Fields: &jira.IssueFields{
                        Project: jira.Project{
                                Key: "ESC",
                        },
                        Type: jira.IssueType{
                                Name: "Investigation",
                        },
                        Priority: &jira.Priority{
                                Name: "NORMAL - P3",
                        },
                        Assignee: &jira.User{
                                Name: "xxxxxx",
                        },
                        Components: []*jira.Component{
                                {
                                        Name: "xxxxxx",
                                },
                        },
                        Description: jiraDescription,
                        Summary:     jiraSummary,
                        Unknowns:    customfields,
                },
        }
        issue, _, err := client.Issue.Create(&i)
        if err != nil {
                fmt.Println(err)
        }
        ticket = issue.Key
        fmt.Println(ticket)
        fmt.Println("Done")
        return ticket
}

// SlackHandle struct
type SlackHandle struct {
        Client    *slack.Client
        BotID     string
        ChannelID string
}

// GetSlackHandle get slack handle
func GetSlackHandle() (s *SlackHandle) {
        client := slack.New("xxxxxx", slack.OptionDebug(true))
        s = &SlackHandle{
                Client:    client,
                BotID:     "xxxxxx",
                ChannelID: "xxxxxx",
        }
        return s
}

// Post Channel Info
func (s *SlackHandle) PostChannelInfo(msg string) {
        s.Client.PostMessage("xxxxxx", slack.MsgOptionText(msg, false))
}

func AxcChannelPost(checkTime string, a ...interface{}) {
        var ticket string
        if len(a) > 0 {
                ticket = fmt.Sprintf("%s", a[0])
        } else {
                ticket = "NA"
        }
        postMsg := fmt.Sprintf("xxxxxx", strings.TrimSpace(checkTime), ticket)
        GetSlackHandle().PostChannelInfo(postMsg)
}

func SocChannelPost(checkTime string, a ...interface{}) {
        var ticket string
        if len(a) > 0 {
                ticket = fmt.Sprintf("%s", a[0])
        } else {
                ticket = "NA"
        }
        postMsg := fmt.Sprintf("xxxxxx", strings.TrimSpace(checkTime), ticket)
        GetSlackHandle().PostChannelInfo(postMsg)
}

// PostInfoAttach slack info with attachment
func (s *SlackHandle) PostInfoAttach(noteMsg string, attachments slack.Attachment) {
        s.Client.PostMessage("xxxxxx", slack.MsgOptionText(noteMsg, false), slack.MsgOptionAttachments(attachments), slack.MsgOptionAsUser(true))
}

// AxcSlackHandle for AX slack handle
func AxcSlackHandle(checkTime string, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList []map[string]string, actionCode int, a ...interface{}) {
        logrus.Info("[QUERY] AxcSlackHandle Start...")
        var ticket string
        if len(a) > 0 {
                ticket = fmt.Sprintf("%s", a[0])
        } else {
                ticket = "NA"
        }
        slackInfoMap := map[int]map[string]string{
                0:  {"barColor": "#66cc00", "postMessage": "xxxxxx"},
                1:  {"barColor": "#ff9d00", "postMessage": "xxxxxx"},
                2:  {"barColor": "#ff9d00", "postMessage": "xxxxxx"},
                11: {"barColor": "#cc0000", "postMessage": "xxxxxx"},
                22: {"barColor": "#cc0000", "postMessage": "xxxxxx"},
        }
        fmt.Println(slackInfoMap)
        attachementFields := []slack.AttachmentField{
                {
                        Title: "*PROCESSED*",
                        Value: strconv.Itoa(len(campaignProcessedList)),
                        Short: true,
                },
                {
                        Title: "*FAIL*",
                        Value: strconv.Itoa(len(campaignFailList)),
                        Short: true,
                },
                {
                        Title: "*NOT COMPLETE*",
                        Value: strconv.Itoa(len(campaignNotCompleteList)),
                        Short: true,
                },
                {
                        Title: "*NOT POST*",
                        Value: strconv.Itoa(len(campaignNotPostList)),
                        Short: true,
                },
        }
        attachments := slack.Attachment{
                Color:      slackInfoMap[actionCode]["barColor"],
                Title:      fmt.Sprintf("xxxxxx", strings.TrimSpace(checkTime)),
                TitleLink:  "https://xxxxxx",
                Text:       slackInfoMap[actionCode]["postMessage"],
                CallbackID: "xxxxxx",
                Fields:     attachementFields,
        }
        GetSlackHandle().PostInfoAttach("xxxxxx", attachments)
}

func SocSlackHandle(checkTime string, campaignFailList, campaignNotCompleteList, campaignNotPostList, campaignProcessedList []map[string]string, actionCode int, a ...interface{}) {
        logrus.Info("[QUERY] SocSlackHandle Start...")
        var ticket string
        if len(a) > 0 {
                ticket = fmt.Sprintf("%s", a[0])
        } else {
                ticket = "NA"
        }
        slackInfoMap := map[int]map[string]string{
                0:  {"barColor": "#66cc00", "postMessage": "xxxxxx"},
                1:  {"barColor": "#ff9d00", "postMessage": "xxxxxx"},
                2:  {"barColor": "#ff9d00", "postMessage": "xxxxxx"},
                11: {"barColor": "#cc0000", "postMessage": "xxxxxx"},
                22: {"barColor": "#cc0000", "postMessage": "xxxxxx"},
        }
        fmt.Println(slackInfoMap)
        attachementFields := []slack.AttachmentField{
                {
                        Title: "*PROCESSED*",
                        Value: strconv.Itoa(len(campaignProcessedList)),
                        Short: true,
                },
                {
                        Title: "*FAIL*",
                        Value: strconv.Itoa(len(campaignFailList)),
                        Short: true,
                },
                {
                        Title: "*NOT COMPLETE*",
                        Value: strconv.Itoa(len(campaignNotCompleteList)),
                        Short: true,
                },
                {
                        Title: "*NOT POST*",
                        Value: strconv.Itoa(len(campaignNotPostList)),
                        Short: true,
                },
        }
        attachments := slack.Attachment{
                Color:      slackInfoMap[actionCode]["barColor"],
                Title:      fmt.Sprintf("xxxxxx", strings.TrimSpace(checkTime)),
                TitleLink:  "https://xxxxxx",
                Text:       slackInfoMap[actionCode]["postMessage"],
                CallbackID: "xxxxxx",
                Fields:     attachementFields,
        }
        GetSlackHandle().PostInfoAttach("xxxxxx", attachments)
}

func cronJobGo() {
        cronMan := cron.New()
        //cronJobAxc := "20 */1 * * * mon,tue,wed,thu,fri"
        //cronJobSoc := "1 */1 * * * mon,tue,wed,thu,fri"
        cronJobAxc := "1 1 13,14 * * mon,tue,wed,thu,fri"
        cronJobSoc := "1 15,25 14 * * mon,tue,wed,thu,fri,sun"
        //cronJobAxc := "1,10,20,30,40,50 * * * mon,tue,wed,thu,fri"
        //cronJobSoc := "5,15,25,35,45,55 * * * mon,tue,wed,thu,fri"
        cronMan.AddFunc(cronJobAxc, goQueryAxc)
        cronMan.AddFunc(cronJobSoc, goQuerySoc)
        cronMan.Start()
        defer cronMan.Stop()
        select {} // to keep main func stay running at the background without quit
}

func getJiraAuth() (jiraUser, jiraPsw []byte) {
        logrus.Info("Get Jira Auth Start...")
        // variableIdentifier
        varJiraUserIdentifier := "xxxxxx"
        varJiraPswIdentifier := "xxxxxx"

        config := conjurapi.Config{
                Account:      os.Getenv("CONJUR_ACCOUNT"),
                ApplianceURL: os.Getenv("CONJUR_APPLIANCE_URL"),
        }

        conjurClient, err := conjurapi.NewClientFromKey(config,
                authn.LoginPair{
                        Login:  os.Getenv("CONJUR_LOGIN"),
                        APIKey: os.Getenv("CONJUR_APIKEY"),
                },
        )
        if err != nil {
                fmt.Println(err)
        }

        jiraUserRes, _ := conjurClient.RetrieveSecretReader(varJiraUserIdentifier)
        jiraPswRes, _ := conjurClient.RetrieveSecretReader(varJiraPswIdentifier)

        jiraUser, _ = conjurapi.ReadResponseBody(jiraUserRes)
        jiraPsw, _ = conjurapi.ReadResponseBody(jiraPswRes)

        fmt.Println(string(jiraUser))
        fmt.Println(string(jiraPsw))

        logrus.Info("Get Jira Auth Done...")

        return jiraUser, jiraPsw
}

func main() {
        //goQueryAxc()
        //goQuerySoc()
        cronJobGo()
}
