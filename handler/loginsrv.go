package handler

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"time"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/crypto"
	"github.com/ory/hydra-client-go/client"
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"go-micro.dev/v4/cache"
	"go-micro.dev/v4/errors"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/util/log"

	pb "loginsrv/proto"
)

// this code added to create a struct that supports io.WriterColoser to get the image bytes
// and create custom writer for github.com/yeqown/go-qrcode/v2 project
type bufferAdaptor struct {
	*bytes.Buffer
}

func (b bufferAdaptor) Close() error {
	return nil
}

func (b bufferAdaptor) Write(p []byte) (int, error) {
	return b.Buffer.Write(p)
}

type Loginsrv struct {
	hydra *client.OryHydra
	salt  string
	cache cache.Cache
}

// Return a new handler
func New(
	hydra *client.OryHydra,
	salt string) *Loginsrv {
	c := cache.NewCache()
	return &Loginsrv{hydra, salt, c}
}

func (e *Loginsrv) CheckRegister(ctx context.Context, req *pb.CheckRequest, rsp *pb.CheckResponse) error {
	challenge := req.Challenge
	identifier := req.Identifire
	session := e.computeSession(challenge, identifier)
	if req.Session != session.Hex() {
		logger.Errorf("session is wrong")

		return nil
	}

	coreID, _, err := e.cache.Get(ctx, req.Session)
	if err != nil || coreID != nil {
		logger.Errorf("get cache error", err.Error())
		return err
	}

	if !req.Accept {
		redirect, err := e.handleReject(req.Challenge)
		if err != nil {
			logger.Errorf("handleReject error", err.Error())
			return err
		}
		rsp.Redirect = redirect
		return nil
	}

	subject := fmt.Sprintf("coreid:%s", coreID)

	RedirectUrl, err := e.handleAccept(req.Challenge, subject, req.Remember)
	if err != nil {
		logger.Errorf("handleAccept error", err.Error())
		return err
	}

	rsp.Redirect = RedirectUrl
	return nil

}
func (e *Loginsrv) Register(ctx context.Context, req *pb.RegisterRequest, rsp *pb.RegisterResponse) error {
	pub, err := crypto.SigToPub(
		crypto.Keccak256(req.Session),
		req.Signature,
	)
	if err != nil {
		logger.Errorf("handleAccept error", err.Error())
		return err
	}
	coreID := crypto.PubkeyToAddress(*pub)

	if coreID != common.BytesToAddress(req.CoreID) {
		logger.Errorf("coreID is not equal")
		return errors.Forbidden("1", "coreID is not equal")
	}

	//in the cache part we need to store session and redirect
	e.cache.Put(ctx, string(req.Session), coreID.Hex(), 10*time.Minute)

	return nil
}

// Call is a single request handler called via client.Call or the generated client code
func (e *Loginsrv) QrCode(ctx context.Context, req *pb.QrCodeRequest, rsp *pb.QrCodeResponse) error {
	log.Info("Received Loginsrv.QrCode request")
	// TODO :: if challenge was not exist throw an error
	login, err := e.hydra.Admin.GetLoginRequest(&admin.GetLoginRequestParams{
		LoginChallenge: req.Challenge,
		Context:        ctx,
	})

	if err != nil {
		return err
	}

	if *login.GetPayload().Skip {
		// how it should work
		// handler.handleAccept(w, req, challenge, *login.Subject, true)
		return nil
	}
	identifier, err := e.generateIdentifier()
	if err != nil {
		// handler.ErrorW(w, http.StatusInternalServerError, "Internal Server Error", "could not generate login identifier: "+err.Error(),
		// 	middlewares.RequestID, requestID)
		return err
	}
	session := e.computeSession(req.Challenge, identifier).Hex()

	queries := url.Values{}

	queries.Add("session", session)
	link := "corepass://corepass.net/login?" + queries.Encode()

	qrc, err := qrcode.New(link)
	if err != nil {
		panic(err)
	}

	b := bufferAdaptor{Buffer: bytes.NewBuffer(nil)}

	writer := standard.NewWithWriter(b,
		standard.WithLogoImageFilePNG("./assets/Group_16454.png"),
	)

	err = qrc.Save(writer)
	if err != nil {
		panic(err)
	}

	qrStr := base64.StdEncoding.EncodeToString(b.Bytes())
	rsp.Challenge = req.Challenge
	rsp.Identifier = identifier
	rsp.Session = session
	rsp.Link = link
	rsp.Qrcode = qrStr
	return nil
}

func (e *Loginsrv) generateIdentifier() (string, error) {
	identifier := make([]byte, 32)
	if _, err := rand.Read(identifier); err != nil {
		return "", err
	}

	return hex.EncodeToString(identifier), nil
}

func (e *Loginsrv) computeSession(challenge, identifier string) common.Hash {
	data := []byte(challenge)
	data = append(data, []byte(identifier)...)
	data = append(data, []byte(e.salt)...)
	return crypto.Keccak256Hash(data)
}

// logger.Infof("Received Loginsrv.Call request: %v", req)

// HANDLE ACCEPT
func (e *Loginsrv) handleAccept(
	challenge string,
	subject string,
	remember bool,
) (string, error) {
	// requestID := req.Context().Value(middlewares.RequestID)
	accept, err := e.hydra.Admin.AcceptLoginRequest(&admin.AcceptLoginRequestParams{
		Body: &models.AcceptLoginRequest{
			Subject:  &subject,
			Remember: remember,
		},
		LoginChallenge: challenge,
		// Context:        req.Context(),
	})
	if err != nil {
		// TODO :: LOG
		// handler.ErrorW(w, http.StatusInternalServerError, "Internal Server Error", "could not accept login request: "+err.Error(),
		// 	middlewares.RequestID, requestID)
		return "", err
	}

	payload := accept.GetPayload()
	return *payload.RedirectTo, nil
}

func (e *Loginsrv) handleReject(
	challenge string,
) (string, error) {

	reject, err := e.hydra.Admin.RejectLoginRequest(&admin.RejectLoginRequestParams{
		Body: &models.RejectRequest{
			Error: "access_denied",
		},
		LoginChallenge: challenge,
		// Context:        req.Context(),
	})
	if err != nil {
		// TODO LOG

		return "", err
		// handler.ErrorW(w, http.StatusInternalServerError, "Internal Server Error", "could not reject login request: "+err.Error(),
		// 	middlewares.RequestID, requestID)
		// return
	}

	payload := reject.GetPayload()
	return *payload.RedirectTo, nil
}
