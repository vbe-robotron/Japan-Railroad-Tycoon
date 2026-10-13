// Package api は HTTP ハンドラと API クライアントを提供する。
//
// net/http の標準ライブラリのみを用いてエンドポイントを公開し、
// hub を介してゲームを操作する。CLI などのクライアント側で
// 利用する api.Client もここに実装する。
package api
