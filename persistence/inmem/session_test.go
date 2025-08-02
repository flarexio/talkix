package inmem

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/flarexio/talkix/llm/message"
	"github.com/flarexio/talkix/session"
)

type sessionRepoTestSuite struct {
	suite.Suite
	sessions session.Repository
	session  *session.Session
}

func (suite *sessionRepoTestSuite) SetupTest() {
	sessions := NewSessionRepository()

	s := session.NewSession("test-user")
	sessions.Save(s)

	suite.sessions = sessions
	suite.session = s
}

func (suite *sessionRepoTestSuite) TestFind() {
	session, err := suite.sessions.Find(suite.session.ID)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.Equal(suite.session.ID, session.ID)
	suite.Equal(suite.session.UserID, session.UserID)
	suite.Equal(len(suite.session.Conversations), len(session.Conversations))
}

func (suite *sessionRepoTestSuite) TestSave() {
	s := suite.session

	conversation := session.NewConversation()
	conversation.AddMessage(message.SystemMessage("You are a helpful assistant."))
	conversation.AddMessage(message.HumanMessage("What is the capital of France?"))
	conversation.AddMessage(message.AIMessage("The capital of France is Paris."))
	conversation.SetIO(
		"What is the capital of France?",
		"The capital of France is Paris.",
	)
	s.AddConversation(conversation)

	err := suite.sessions.Save(suite.session)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	session, err := suite.sessions.Find(s.ID)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.Len(session.Conversations, 1)
}

func TestSessionRepoTestSuite(t *testing.T) {
	suite.Run(t, new(sessionRepoTestSuite))
}
