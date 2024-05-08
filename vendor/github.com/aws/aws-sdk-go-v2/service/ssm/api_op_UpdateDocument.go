// Code generated by smithy-go-codegen DO NOT EDIT.

package ssm

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Updates one or more values for an SSM document.
func (c *Client) UpdateDocument(ctx context.Context, params *UpdateDocumentInput, optFns ...func(*Options)) (*UpdateDocumentOutput, error) {
	if params == nil {
		params = &UpdateDocumentInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "UpdateDocument", params, optFns, c.addOperationUpdateDocumentMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*UpdateDocumentOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type UpdateDocumentInput struct {

	// A valid JSON or YAML string.
	//
	// This member is required.
	Content *string

	// The name of the SSM document that you want to update.
	//
	// This member is required.
	Name *string

	// A list of key-value pairs that describe attachments to a version of a document.
	Attachments []types.AttachmentsSource

	// The friendly name of the SSM document that you want to update. This value can
	// differ for each version of the document. If you don't specify a value for this
	// parameter in your request, the existing value is applied to the new document
	// version.
	DisplayName *string

	// Specify the document format for the new document version. Systems Manager
	// supports JSON and YAML documents. JSON is the default format.
	DocumentFormat types.DocumentFormat

	// The version of the document that you want to update. Currently, Systems Manager
	// supports updating only the latest version of the document. You can specify the
	// version number of the latest version or use the $LATEST variable.
	//
	// If you change a document version for a State Manager association, Systems
	// Manager immediately runs the association unless you previously specifed the
	// apply-only-at-cron-interval parameter.
	DocumentVersion *string

	// Specify a new target type for the document.
	TargetType *string

	// An optional field specifying the version of the artifact you are updating with
	// the document. For example, 12.6. This value is unique across all versions of a
	// document, and can't be changed.
	VersionName *string

	noSmithyDocumentSerde
}

type UpdateDocumentOutput struct {

	// A description of the document that was updated.
	DocumentDescription *types.DocumentDescription

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationUpdateDocumentMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpUpdateDocument{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpUpdateDocument{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "UpdateDocument"); err != nil {
		return fmt.Errorf("add protocol finalizers: %v", err)
	}

	if err = addlegacyEndpointContextSetter(stack, options); err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = addClientRequestID(stack); err != nil {
		return err
	}
	if err = addComputeContentLength(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = addComputePayloadSHA256(stack); err != nil {
		return err
	}
	if err = addRetry(stack, options); err != nil {
		return err
	}
	if err = addRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = addRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addClientUserAgent(stack, options); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addSetLegacyContextSigningOptionsMiddleware(stack); err != nil {
		return err
	}
	if err = addOpUpdateDocumentValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opUpdateDocument(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = addRecursionDetection(stack); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	if err = addDisableHTTPSMiddleware(stack, options); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opUpdateDocument(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "UpdateDocument",
	}
}
