# Project Ambiguities and Questions

This document outlines the most ambiguous aspects of the Document Split + Review Backend project, based on the current implementation and requirements.

## 1. Error Handling for Empty IDs
**Ambiguity**: The requirements don't specify how to handle empty IDs in endpoints.
**Example**: 
- Option 1: Return 400 with "ID is required" (treat as bad request)
- Option 2: Return 404 with "not found" (treat as missing resource)
- Option 3: Return 204 (treat as successful no-op)
**What Caused It**: Different handlers implemented different approaches, leading to inconsistent behavior.

## 2. Response Status Codes
**Ambiguity**: The requirements don't explicitly specify which HTTP status codes to use.
**Example**:
- Option 1: Use 201 for document creation, 200 for updates
- Option 2: Use 200 for all successful operations
- Option 3: Use 204 for all operations that don't return content
**What Caused It**: REST best practices suggest different status codes, but requirements didn't specify which to follow.

## 3. Document Structure in LoadSplit Response
**Ambiguity**: The requirements show a sample JSON with `page_urls` array containing PNG files, but this doesn't match the actual document processing needs.
**Example**:
- Requirements show: `"page_urls": ["page_1.png", "page_2.png", "page_3.png"]`
- Actual Implementation:
  ```go
  type Page struct {
      ID         string  // e.g. "page_1"
      SplitID    string  // ID of the split this page belongs to
      DocumentID *string // nil if unassigned
      PageNumber int     // its position in the original PDF
      URL        string  // thumbnail / image URL for preview
  }
  ```
- Key Points:
  1. Content Storage:
     - Page content is stored on the filesystem, not in the database
     - Database only stores metadata and references
     - Content path is derived from SplitID and PageNumber
  2. Preview URLs:
     - Thumbnail URLs are stored for preview purposes
     - These are separate from the actual content files
  3. Page Organization:
     - Pages maintain their original PDF order via PageNumber
     - Can be assigned to documents or left unassigned
     - Document tracks its page range (StartPage to EndPage)

**What Caused It**: 
- Requirements example used PNG files without explaining storage strategy
- No clear separation between content storage and metadata
- Unclear distinction between preview thumbnails and actual content

**Impact**:
- Initial implementation had to be revised to separate content storage
- Database schema needed to be updated to remove content field
- File system storage strategy needed to be implemented
- Preview/thumbnail system needed to be properly integrated

## 4. Page Movement Logic
**Ambiguity**: The requirements didn't fully specify the rules and constraints for page movement operations.
**Example**:
- Current Implementation:
  ```go
  type MovePagesRequest struct {
      SplitID        string   // ID of the split containing the documents
      FromDocumentID string   // source document ID
      ToDocumentID   string   // destination document ID
      PageIDs        []string // list of page IDs to move
  }
  ```
- Key Ambiguities:
  1. Page Ordering:
     - Should pages maintain their original PDF order?
     - Should they be reordered based on the target document's structure?
     - Current: We maintain original page numbers but allow reordering
  2. Document Continuity:
     - Should documents maintain continuous page ranges?
     - What happens to document metadata (StartPage/EndPage) after moves?
     - Current: We update StartPage/EndPage but don't enforce continuity
  3. Unassigned Pages:
     - Can pages be moved to an unassigned state?
     - How are unassigned pages tracked and managed?
     - Current: We support unassigned pages but don't expose this in the API
  4. Validation Rules:
     - What constraints should be placed on page movement?
     - How to handle invalid moves (e.g., moving non-existent pages)?
     - Current: Basic validation but no business rule enforcement

**What Caused It**: 
- Lack of clear business rules for page organization
- No specification of document continuity requirements
- Unclear requirements for page ordering and validation
- Missing rules for handling unassigned pages

**Impact**:
- Inconsistent page ordering across documents
- Potential gaps in page ranges
- Limited support for unassigned pages
- Basic validation without business rule enforcement

## 5. Document Finalization Rules
**Ambiguity**: The requirements don't specify finalization criteria.
**Example**:
- Option 1: Allow finalization with unassigned pages
- Option 2: Require all pages to be assigned
- Option 3: Allow finalization with warnings
**What Caused It**: No clear definition of what constitutes a "finalized" state.

## 6. Authorization and Access Control
**Ambiguity**: The requirements mention "basic authorization" but don't specify details.
**Example**:
- Option 1: Simple API key per client
- Option 2: JWT-based authentication
- Option 3: OAuth2 with client credentials
**What Caused It**: "Basic authorization" is too vague to implement consistently.

## 7. Document Classification
**Ambiguity**: The requirements don't specify classification rules or valid values.
**Example**:
- Option 1: Free-form text classification
- Option 2: Predefined list (W-2, 1099, etc.)
- Option 3: Hierarchical classification system
**What Caused It**: No taxonomy or validation rules provided for document types.

## 8. File Naming and Storage
**Ambiguity**: The requirements don't specify file naming conventions or storage rules.
**Example**:
- Option 1: Use original filenames
- Option 2: Generate UUID-based names
- Option 3: Use client_id/document_id based names
**What Caused It**: No guidelines for file management and storage.

## 9. Error Recovery and Retry Logic
**Ambiguity**: The requirements don't specify how to handle failures.
**Example**:
- Option 1: Fail fast, no retries
- Option 2: Retry with exponential backoff
- Option 3: Queue for manual review
**What Caused It**: No specification of failure handling strategies.

## 10. Observability Requirements
**Ambiguity**: The requirements mention observability but don't specify metrics or logging.
**Example**:
- Option 1: Basic error logging only
- Option 2: Full request/response logging
- Option 3: Structured logging with metrics
**What Caused It**: "Observability" mentioned without specific requirements.

## Recommendations

1. **Documentation Updates**
   - Add explicit status code requirements
   - Define error handling patterns
   - Specify validation rules

2. **Implementation Guidelines**
   - Create consistent error handling patterns
   - Define clear response structures
   - Establish naming conventions

3. **Testing Requirements**
   - Add test cases for edge cases
   - Define expected behaviors
   - Document test scenarios

4. **Future Considerations**
   - Plan for scalability
   - Consider multi-tenant support
   - Design for extensibility 


# Answers

## 1. Error Handling for Empty IDs
**Ambiguity**: The requirements don't specify how to handle empty IDs in endpoints.
**Solution**: 
- Option 1: Return 400 with "ID is required" (treat as bad request)



## 2. Response Status Codes
**Ambiguity**: The requirements don't explicitly specify which HTTP status codes to use.
**Solution**:
- Option 1: Use 201 for document creation, 200 for updates



## 3. Document Structure in LoadSplit Response
**Ambiguity**: The requirements show a sample JSON with `page_urls` array containing PNG files, but this doesn't match the actual document processing needs.
**Example**:
- Requirements show: `"page_urls": ["page_1.png", "page_2.png", "page_3.png"]`
- Actual Implementation:
  ```go
  type Page struct {
      ID         string  // e.g. "page_1"
      SplitID    string  // ID of the split this page belongs to
      DocumentID *string // nil if unassigned
      PageNumber int     // its position in the original PDF
      URL        string  // thumbnail / image URL for preview
  }
  ```
- Key Points:
  1. Content Storage:
     - Page content is stored on the filesystem, not in the database
     - Database only stores metadata and references
     - Content path is derived from SplitID and PageNumber
  2. Preview URLs:
     - Thumbnail URLs are stored for preview purposes
     - These are separate from the actual content files
  3. Page Organization:
     - Pages maintain their original PDF order via PageNumber
     - Can be assigned to documents or left unassigned
     - Document tracks its page range (StartPage to EndPage)

**What Caused It**: 
- Requirements example used PNG files without explaining storage strategy
- No clear separation between content storage and metadata
- Unclear distinction between preview thumbnails and actual content

**Impact**:
- Initial implementation had to be revised to separate content storage
- Database schema needed to be updated to remove content field
- File system storage strategy needed to be implemented
- Preview/thumbnail system needed to be properly integrated