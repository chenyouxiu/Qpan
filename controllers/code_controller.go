package controllers

import (
	"Qpan/utils"
	"bytes"
	"encoding/base64"
	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
	"net/http"
)

type NewCaptchaResponse struct {
	CaptchaID string `json:"captcha_id"` // CAPTCHA 的唯一标识符
	Image     string `json:"image"`      // CAPTCHA 图像的 Base64 编码
}

type VerifyCaptchaRequest struct {
	CaptchaID string `json:"captcha_id"`       // CAPTCHA 的唯一标识符
	Solution  string `json:"captcha_solution"` // 用户输入的 CAPTCHA 解决方案
}

func GetCaptcha(c *gin.Context) {
	captchaID := captcha.New()
	var buf bytes.Buffer

	if err := captcha.WriteImage(&buf, captchaID, captcha.StdWidth, captcha.StdHeight); err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Code: utils.ERROR,
			Msg:  "生成验证码失败",
			Data: nil,
		})
		return
	}

	encodedImage := base64.StdEncoding.EncodeToString(buf.Bytes())

	responseData := NewCaptchaResponse{
		CaptchaID: captchaID,
		Image:     encodedImage,
	}
	c.JSON(http.StatusOK, utils.Response{
		Code: 200,
		Msg:  "生成 CAPTCHA 成功",
		Data: responseData,
	})
}

func VerifyCaptcha(c *gin.Context) {
	var req VerifyCaptchaRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Code: utils.INVALID_PARAMS,
			Msg:  "请求参数错误",
			Data: nil,
		})
		return
	}

	if captcha.VerifyString(req.CaptchaID, req.Solution) {
		c.JSON(http.StatusOK, utils.Response{
			Code: 200,
			Msg:  "验证码正确",
			Data: nil,
		})
	} else {
		c.JSON(http.StatusBadRequest, utils.Response{
			Code: 400,
			Msg:  "验证码错误",
			Data: nil,
		})
	}
}
